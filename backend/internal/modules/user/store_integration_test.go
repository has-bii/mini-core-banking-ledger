//go:build integration

package user_test

import (
	"errors"
	"testing"

	"backend/internal/modules/user"
	"backend/internal/testutil/testdb"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestStorePersistsUser(t *testing.T) {
	t.Parallel()
	store := user.NewStore(testdb.Open(t))
	created, err := store.Create(t.Context(), "Alice", "alice@example.com")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := uuid.Parse(created.ID); err != nil || created.Name != "Alice" || created.Email != "alice@example.com" || created.Image != nil || created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Fatalf("unexpected created user: %+v", created)
	}
	byID, err := store.FindByID(t.Context(), created.ID)
	if err != nil || byID != created {
		t.Fatalf("FindByID() = %+v, %v; want %+v", byID, err, created)
	}
	byEmail, err := store.FindByEmail(t.Context(), created.Email)
	if err != nil || byEmail != created {
		t.Fatalf("FindByEmail() = %+v, %v; want %+v", byEmail, err, created)
	}
	duplicate, err := store.Create(t.Context(), "Other", created.Email)
	if !errors.Is(err, user.ErrEmailTaken) || duplicate != (user.User{}) {
		t.Fatalf("duplicate Create() = %+v, %v; want empty user and ErrEmailTaken", duplicate, err)
	}
	unchanged, err := store.FindByID(t.Context(), created.ID)
	if err != nil || unchanged != created {
		t.Fatal("duplicate creation changed the existing user")
	}
}

func TestStoreUserErrors(t *testing.T) {
	t.Parallel()
	store := user.NewStore(testdb.Open(t))
	if _, err := store.FindByID(t.Context(), uuid.NewString()); !errors.Is(err, user.ErrNotFound) {
		t.Errorf("missing ID error = %v", err)
	}
	if _, err := store.FindByEmail(t.Context(), "missing@example.com"); !errors.Is(err, user.ErrNotFound) {
		t.Errorf("missing email error = %v", err)
	}
	_, err := store.FindByID(t.Context(), "not-a-uuid")
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "22P02" || errors.Is(err, user.ErrNotFound) {
		t.Errorf("unexpected database failure must preserve its cause: %v", err)
	}
}
