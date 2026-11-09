//go:build integration

package testdb_test

import (
	"errors"
	"testing"

	"backend/internal/modules/user"
	"backend/internal/testutil/testdb"
)

func TestOpenIsolatesMigratesAndCleansUp(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://ignored:ignored@127.0.0.1:1/ignored")
	pool := testdb.Open(t)
	store := user.NewStore(pool)
	original, err := store.Create(t.Context(), "Original", "alice@example.com")
	if err != nil {
		t.Fatalf("migrations were not applied: %v", err)
	}
	var childSchema string
	t.Run("isolated schema", func(t *testing.T) {
		isolated := testdb.Open(t)
		if err := isolated.QueryRow(t.Context(), `SELECT current_schema()`).Scan(&childSchema); err != nil {
			t.Fatal(err)
		}
		other := user.NewStore(isolated)
		if _, err := other.FindByEmail(t.Context(), original.Email); !errors.Is(err, user.ErrNotFound) {
			t.Fatalf("isolated schema can see another test's user: %v", err)
		}
		if _, err := other.Create(t.Context(), "Isolated", original.Email); err != nil {
			t.Fatalf("isolated schema cannot reuse another test's email: %v", err)
		}
	})
	if childSchema == "" {
		t.Fatal("child schema was not created")
	}
	var exists bool
	if err := pool.QueryRow(t.Context(), `SELECT EXISTS (SELECT 1 FROM pg_namespace WHERE nspname = $1)`, childSchema).Scan(&exists); err != nil {
		t.Fatal(err)
	}
	if exists {
		t.Error("test cleanup left its schema behind")
	}
	got, err := store.FindByEmail(t.Context(), original.Email)
	if err != nil || got != original {
		t.Fatalf("child cleanup affected another test: %+v, %v", got, err)
	}
}
