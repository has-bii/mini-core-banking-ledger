//go:build integration

package account_test

import (
	"backend/internal/modules/account"
	"backend/internal/modules/user"
	"backend/internal/platform/validation"
	"backend/internal/testutil"
	"errors"
	"reflect"
	"testing"
)

func TestStore_GetByID(t *testing.T) {
	t.Parallel()
	pool := testutil.Postgres(t)
	users := user.NewStore(pool)
	owner, err := users.Create(t.Context(), "Alice", "alice@example.com", "stored-password-hash")
	if err != nil {
		t.Fatal(err)
	}
	other, err := users.Create(t.Context(), "Bob", "bob@example.com", "stored-password-hash")
	if err != nil {
		t.Fatal(err)
	}
	store := account.NewStore(pool)
	accounts, err := store.ListByUserID(t.Context(), owner.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(accounts) != 1 {
		t.Fatalf("owner accounts = %d, want 1", len(accounts))
	}
	want := accounts[0]

	for _, tt := range []struct {
		name   string
		id     string
		userID string
		err    error
	}{
		{name: "returns owned account", id: want.ID, userID: owner.ID},
		{name: "missing account", id: "00000000-0000-0000-0000-000000000000", userID: owner.ID, err: account.ErrNotFound},
		{name: "another user's account", id: want.ID, userID: other.ID, err: account.ErrNotFound},
		{name: "missing user", id: want.ID, userID: "00000000-0000-0000-0000-000000000000", err: account.ErrNotFound},
		{name: "invalid account ID", id: "invalid", userID: owner.ID, err: validation.ErrInvalidUUID},
		{name: "empty account ID", userID: owner.ID, err: validation.ErrInvalidUUID},
		{name: "invalid user ID", id: want.ID, userID: "invalid", err: validation.ErrInvalidUUID},
		{name: "empty user ID", id: want.ID, err: validation.ErrInvalidUUID},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.GetByID(t.Context(), tt.id, tt.userID)
			if !errors.Is(err, tt.err) {
				t.Fatalf("GetByID() error = %v, want %v", err, tt.err)
			}
			if tt.err != nil {
				if !reflect.DeepEqual(got, account.Account{}) {
					t.Errorf("GetByID() returned account on error: %+v", got)
				}
				return
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("GetByID() = %+v, want %+v", got, want)
			}
		})
	}

	t.Run("database error", func(t *testing.T) {
		pool.Close()
		got, err := store.GetByID(t.Context(), want.ID, owner.ID)
		if err == nil || errors.Is(err, account.ErrNotFound) || errors.Is(err, validation.ErrInvalidUUID) {
			t.Fatalf("GetByID() error = %v, want database error", err)
		}
		if !reflect.DeepEqual(got, account.Account{}) {
			t.Errorf("GetByID() returned account on error: %+v", got)
		}
	})
}

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
