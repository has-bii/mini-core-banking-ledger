//go:build integration

package user_test

import (
	"backend/internal/modules/user"
	"backend/internal/testutil"
	"errors"
	"testing"
)

func TestStore_UserPersistence(t *testing.T) {
	t.Parallel()
	store := user.NewStore(testutil.Postgres(t))
	created, err := store.Create(t.Context(), "Alice", "alice@example.com", "stored-password-hash")
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == "" || created.Name != "Alice" || created.Email != "alice@example.com" || created.Image != nil || created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Fatalf("Create() = %+v, want populated user", created)
	}
	byID, err := store.FindByID(t.Context(), created.ID)
	if err != nil || byID != created {
		t.Fatalf("FindByID() = %+v, %v; want %+v", byID, err, created)
	}
	byEmail, err := store.FindByEmail(t.Context(), created.Email)
	if err != nil || byEmail != created {
		t.Fatalf("FindByEmail() = %+v, %v; want %+v", byEmail, err, created)
	}
	if _, err := store.Create(t.Context(), "Another Alice", created.Email, "stored-password-hash"); !errors.Is(err, user.ErrEmailTaken) {
		t.Fatalf("duplicate email error = %v, want ErrEmailTaken", err)
	}
	if _, err := store.FindByID(t.Context(), "00000000-0000-0000-0000-000000000000"); !errors.Is(err, user.ErrNotFound) {
		t.Fatalf("missing ID error = %v, want ErrNotFound", err)
	}
	if _, err := store.FindByEmail(t.Context(), "missing@example.com"); !errors.Is(err, user.ErrNotFound) {
		t.Fatalf("missing email error = %v, want ErrNotFound", err)
	}
}
