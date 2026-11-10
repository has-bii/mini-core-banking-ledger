//go:build integration

package account_test

import (
	"backend/internal/modules/account"
	"backend/internal/modules/user"
	"backend/internal/platform/validation"
	"backend/internal/testutil"
	"testing"
)

func TestStore_ListAccounts(t *testing.T) {
	t.Parallel()
	pool := testutil.Postgres(t)

	userStore := user.NewStore(pool)

	// Preparing data, create new users
	u, err := userStore.Create(t.Context(), "Alice", "alice@example.com", "stored-password-hash")
	if err != nil {
		t.Fatal(err)
	}
	_, err = userStore.Create(t.Context(), "Bob", "bob@example.com", "stored-password-hash")
	if err != nil {
		t.Fatal(err)
	}

	store := account.NewStore(pool)

	accounts, err := store.ListByUserID(t.Context(), u.ID)
	if err != nil {
		t.Fatal(err)
	}

	// accounts count
	if length := len(accounts); length != 1 {
		t.Fatalf("ListByUserID() length != 1, %d", length)
	}

	got := &accounts[0]

	// Check ownership
	if got.UserID != u.ID {
		t.Errorf("Ownership failed! Expected UserID to be '%v', got '%v'", u.ID, got.UserID)
	}

	// mapping
	if !validation.IsValidUUID(got.ID) {
		t.Errorf("Mapping failed! Expected ID to be valid UUID, got '%v'", got.ID)
	}
	if !validation.IsValidUUID(got.UserID) {
		t.Errorf("Mapping failed! Expected UserID to be valid UUID, got '%v'", got.UserID)
	}
	if got.Name != "Main Account" {
		t.Errorf("Mapping failed! Expected Name to be 'Main Account', got '%v'", got.Name)
	}
	if got.Type != account.Main {
		t.Errorf("Mapping failed! Expected Type to be '%v', got '%v'", account.Main, got.Type)
	}
	if got.SortOrder != 0 {
		t.Errorf("Mapping failed! Expected SortOrder to be 0, got '%v'", got.SortOrder)
	}
	if got.Balance != 0 {
		t.Errorf("Mapping failed! Expected Balance to be 0, got '%v'", got.Balance)
	}
	if got.CreatedAt.IsZero() {
		t.Errorf("Mapping failed! Expected CreatedAt to be populated, got '%v'", got.CreatedAt)
	}
}
