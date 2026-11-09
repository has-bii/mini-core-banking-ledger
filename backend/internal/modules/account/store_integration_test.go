//go:build integration

package account_test

import (
	"errors"
	"testing"

	"backend/internal/modules/account"
	"backend/internal/modules/user"
	"backend/internal/testutil/testdb"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestStorePersistsCredentials(t *testing.T) {
	t.Parallel()
	pool := testdb.Open(t)
	u, err := user.NewStore(pool).Create(t.Context(), "Alice", "alice@example.com")
	if err != nil {
		t.Fatal(err)
	}
	store := account.NewStore(pool)
	created, err := store.Create(t.Context(), u.ID, "hashed-password")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := uuid.Parse(created.ID); err != nil || created.UserID != u.ID || created.Password == nil || *created.Password != "hashed-password" || created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Fatalf("unexpected credentials: %+v", created)
	}
	got, err := store.FindByUserID(t.Context(), u.ID)
	if err != nil || got.ID != created.ID || got.UserID != u.ID || got.Password == nil || *got.Password != *created.Password || !got.CreatedAt.Equal(created.CreatedAt) || !got.UpdatedAt.Equal(created.UpdatedAt) {
		t.Fatalf("credentials did not round-trip: %+v, %v", got, err)
	}
}

func TestStoreAccountErrors(t *testing.T) {
	t.Parallel()
	store := account.NewStore(testdb.Open(t))
	missingID := uuid.NewString()
	if _, err := store.FindByUserID(t.Context(), missingID); !errors.Is(err, account.ErrNotFound) {
		t.Fatalf("missing account error = %v", err)
	}
	_, err := store.Create(t.Context(), missingID, "hashed-password")
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23503" {
		t.Errorf("foreign-key failure must preserve its cause: %v", err)
	}
	_, err = store.FindByUserID(t.Context(), "invalid-uuid")
	if !errors.As(err, &pgErr) || pgErr.Code != "22P02" || errors.Is(err, account.ErrNotFound) {
		t.Errorf("unexpected lookup failure must preserve its cause: %v", err)
	}
}
