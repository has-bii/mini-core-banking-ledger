//go:build integration

package testutil_test

import (
	"backend/internal/testutil"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func TestPostgres_IsolatesSchemasAndConnections(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgresql://invalid:invalid@127.0.0.1:1/invalid")
	t.Setenv("GOOSE_DBSTRING", "postgresql://invalid:invalid@127.0.0.1:1/invalid")
	t.Setenv("PGOPTIONS", "-c search_path=public")
	var schema string
	t.Run("private schema", func(t *testing.T) {
		first := testutil.Postgres(t)
		second := testutil.Postgres(t)
		if err := first.QueryRow(t.Context(), "SELECT current_schema()").Scan(&schema); err != nil {
			t.Fatal(err)
		}
		for _, pool := range []*pgxpool.Pool{first, second} {
			one, err := pool.Acquire(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(one.Release)
			two, err := pool.Acquire(t.Context())
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(two.Release)
			var a, b, path string
			if err := one.QueryRow(t.Context(), "SELECT current_schema(), current_setting('search_path')").Scan(&a, &path); err != nil {
				t.Fatal(err)
			}
			if err := two.QueryRow(t.Context(), "SELECT current_schema()").Scan(&b); err != nil {
				t.Fatal(err)
			}
			if a == "public" || a == "" || a != b || path != a {
				t.Fatalf("connections escaped their schema: %q, %q, search_path=%q", a, b, path)
			}
		}
		if _, err := first.Exec(t.Context(), `INSERT INTO users (name, email, password, "createdAt", "updatedAt") VALUES ('Alice', 'alice@example.com', 'stored-password-hash', now(), now())`); err != nil {
			t.Fatal(err)
		}
		var count int
		if err := second.QueryRow(t.Context(), "SELECT count(*) FROM users").Scan(&count); err != nil || count != 0 {
			t.Fatalf("second schema contains %d users, want none (%v)", count, err)
		}
	})
	pool := testutil.Postgres(t)
	var exists bool
	if err := pool.QueryRow(t.Context(), "SELECT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = $1)", schema).Scan(&exists); err != nil || exists {
		t.Fatalf("schema survived test cleanup: exists=%v (%v)", exists, err)
	}
}
