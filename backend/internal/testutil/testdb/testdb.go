//go:build integration

// Package testdb provides a migrated, isolated PostgreSQL schema per test.
package testdb

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"backend/internal/platform/database"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// Open reads only backend/.env.test, applies migrations, and drops its own schema
// at cleanup. Tests never share tables or modify the database's public schema.
func Open(t *testing.T) *pgxpool.Pool {
	t.Helper()
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("locate backend test configuration")
	}
	root := filepath.Join(filepath.Dir(source), "..", "..", "..")
	env, err := godotenv.Read(filepath.Join(root, ".env.test"))
	if err != nil {
		t.Fatalf("read backend/.env.test: %v", err)
	}
	if env["DATABASE_URL"] == "" {
		t.Fatal("DATABASE_URL is required in backend/.env.test")
	}
	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	admin, err := database.Open(ctx, env["DATABASE_URL"])
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(admin.Close)

	schema := "test_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	quoted := pgx.Identifier{schema}.Sanitize()
	if _, err := admin.Exec(ctx, "CREATE SCHEMA "+quoted); err != nil {
		t.Fatalf("create test schema: %v", err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := admin.Exec(ctx, "DROP SCHEMA "+quoted+" CASCADE"); err != nil {
			t.Errorf("drop test schema: %v", err)
		}
	})

	cfg := admin.Config()
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatalf("open isolated test pool: %v", err)
	}
	t.Cleanup(pool.Close)

	files, err := filepath.Glob(filepath.Join(root, "migrations", "*.sql"))
	if err != nil || len(files) == 0 {
		t.Fatalf("locate SQL migrations: %v", err)
	}
	for _, file := range files {
		body, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		// shortcut: SQL-only Goose Up/Down sections; use Goose's runner if advanced directives or Go migrations are added.
		for _, line := range strings.Split(string(body), "\n") {
			directive := strings.TrimSpace(line)
			if strings.HasPrefix(directive, "-- +goose ") && directive != "-- +goose Up" && directive != "-- +goose Down" {
				t.Fatalf("unsupported migration directive %q in %s; use Goose's runner", directive, file)
			}
		}
		_, up, found := strings.Cut(string(body), "-- +goose Up")
		if !found {
			t.Fatalf("migration %s has no Goose Up section", file)
		}
		up, _, _ = strings.Cut(up, "-- +goose Down")
		if err := database.InTx(ctx, pool, func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, up)
			return err
		}); err != nil {
			t.Fatalf("apply migration %s: %v", filepath.Base(file), err)
		}
	}
	return pool
}
