//go:build integration

package user_test

import (
	"backend/internal/modules/user"
	"backend/internal/testutil"
	"errors"
	"testing"
)

func TestStore_CreateMainAccount(t *testing.T) {
	t.Parallel()
	db := testutil.Postgres(t)
	store := user.NewStore(db)
	created, err := store.Create(t.Context(), "Alice", "alice@example.com", "stored-password-hash")
	if err != nil {
		t.Fatal(err)
	}
	var count int
	if err := db.QueryRow(t.Context(), `SELECT count(*) FROM accounts WHERE user_id = $1`, created.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("account count = %d, want 1", count)
	}
	var name, accountType string
	var sortOrder, balance int64
	if err := db.QueryRow(t.Context(), `SELECT name, type, sort_order, balance FROM accounts WHERE user_id = $1`, created.ID).Scan(&name, &accountType, &sortOrder, &balance); err != nil {
		t.Fatal(err)
	}
	if name != "Main Account" || accountType != "main" || sortOrder != 0 || balance != 0 {
		t.Fatalf("main account = (%q, %q, %d, %d), want (Main Account, main, 0, 0)", name, accountType, sortOrder, balance)
	}
}

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
