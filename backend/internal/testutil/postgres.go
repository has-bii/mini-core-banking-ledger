//go:build integration

package testutil

import (
	"backend/internal/platform/storage/postgres"
	"context"
	"crypto/rand"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// Postgres migrates a private schema and removes only that schema after the test.
func Postgres(t *testing.T) *pgxpool.Pool {
	t.Helper()
	root, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			t.Fatal("cannot locate backend/go.mod")
		}
		root = parent
	}
	values, err := godotenv.Read(filepath.Join(root, ".env.test"))
	if err != nil {
		t.Fatalf("read backend/.env.test: %v", err)
	}
	dsn, err := url.Parse(values["DATABASE_URL"])
	if err != nil || (dsn.Scheme != "postgres" && dsn.Scheme != "postgresql") || dsn.Host == "" || strings.Trim(dsn.Path, "/") == "" {
		t.Fatal("backend/.env.test must contain a PostgreSQL DATABASE_URL with an explicit host and test database")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	admin, err := pgx.Connect(ctx, dsn.String())
	if err != nil {
		t.Fatalf("connect to test database: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := admin.Close(ctx); err != nil {
			t.Errorf("close test admin connection: %v", err)
		}
	})
	schema := "test_" + strings.ToLower(rand.Text())
	identifier := pgx.Identifier{schema}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+identifier); err != nil {
		t.Fatalf("create test schema: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := admin.Exec(ctx, "DROP SCHEMA "+identifier+" CASCADE"); err != nil {
			t.Errorf("drop test schema %s: %v", schema, err)
		}
	})

	query := dsn.Query()
	query.Set("search_path", schema)
	query.Del("options") // Do not let URL startup options override schema isolation.
	dsn.RawQuery = query.Encode()
	cmd := exec.CommandContext(ctx, "goose", "-env", "none", "-dir", filepath.Join(root, "migrations"), "postgres", dsn.String(), "up")
	cmd.Env = []string{"PATH=" + os.Getenv("PATH"), "HOME=" + os.Getenv("HOME"), "NO_COLOR=1"}
	if output, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("apply test migrations (Goose CLI required): %v\n%s", err, output)
	}
	pool, err := postgres.Open(ctx, dsn.String())
	if err != nil {
		t.Fatalf("open isolated test pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
}
