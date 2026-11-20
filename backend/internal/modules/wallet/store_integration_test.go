//go:build integration

package wallet_test

import (
	"backend/internal/modules/user"
	"backend/internal/modules/wallet"
	"backend/internal/testutil"
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
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
	store := wallet.NewStore(pool)
	wallets, err := store.ListByUserID(t.Context(), owner.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(wallets) != 1 {
		t.Fatalf("owner wallets = %d, want 1", len(wallets))
	}
	want := wallets[0]

	for _, tt := range []struct {
		name   string
		id     string
		userID string
		err    error
	}{
		{name: "returns owned wallet", id: want.ID, userID: owner.ID},
		{name: "missing wallet", id: "00000000-0000-0000-0000-000000000000", userID: owner.ID, err: wallet.ErrNotFound},
		{name: "another user's wallet", id: want.ID, userID: other.ID, err: wallet.ErrNotFound},
		{name: "missing user", id: want.ID, userID: "00000000-0000-0000-0000-000000000000", err: wallet.ErrNotFound},
		{name: "invalid wallet ID", id: "invalid", userID: owner.ID, err: wallet.ErrInvalidWalletID},
		{name: "empty wallet ID", userID: owner.ID, err: wallet.ErrInvalidWalletID},
		{name: "invalid user ID", id: want.ID, userID: "invalid", err: wallet.ErrInvalidUserID},
		{name: "empty user ID", id: want.ID, err: wallet.ErrInvalidUserID},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.GetByID(t.Context(), tt.id, tt.userID)
			if !errors.Is(err, tt.err) {
				t.Fatalf("GetByID() error = %v, want %v", err, tt.err)
			}
			if tt.err != nil {
				if !reflect.DeepEqual(got, wallet.Wallet{}) {
					t.Errorf("GetByID() returned wallet on error: %+v", got)
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
		if err == nil || errors.Is(err, wallet.ErrNotFound) || errors.Is(err, wallet.ErrInvalidWalletID) || errors.Is(err, wallet.ErrInvalidUserID) {
			t.Fatalf("GetByID() error = %v, want database error", err)
		}
		if !reflect.DeepEqual(got, wallet.Wallet{}) {
			t.Errorf("GetByID() returned wallet on error: %+v", got)
		}
	})
}

func TestStore_ListByUserID(t *testing.T) {
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

	store := wallet.NewStore(pool)

	wallets, err := store.ListByUserID(t.Context(), u.ID)
	if err != nil {
		t.Fatal(err)
	}

	// wallets count
	if length := len(wallets); length != 1 {
		t.Fatalf("ListByUserID() length != 1, %d", length)
	}

	got := &wallets[0]

	// Check ownership
	if got.UserID != u.ID {
		t.Errorf("Ownership failed! Expected UserID to be '%v', got '%v'", u.ID, got.UserID)
	}

	// mapping
	if uuid.Validate(got.ID) != nil {
		t.Errorf("Mapping failed! Expected ID to be valid UUID, got '%v'", got.ID)
	}
	if uuid.Validate(got.UserID) != nil {
		t.Errorf("Mapping failed! Expected UserID to be valid UUID, got '%v'", got.UserID)
	}
	if got.Name != "Main Wallet" {
		t.Errorf("Mapping failed! Expected Name to be 'Main Wallet', got '%v'", got.Name)
	}
	if got.Type != wallet.Main {
		t.Errorf("Mapping failed! Expected Type to be '%v', got '%v'", wallet.Main, got.Type)
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

	for _, tt := range []struct {
		name   string
		userID string
		err    error
	}{
		{name: "invalid user ID", userID: "invalid", err: wallet.ErrInvalidUserID},
		{name: "empty user ID", err: wallet.ErrInvalidUserID},
		{name: "no matching wallets", userID: "00000000-0000-0000-0000-000000000000"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.ListByUserID(t.Context(), tt.userID)
			if !errors.Is(err, tt.err) || len(got) != 0 {
				t.Errorf("ListByUserID() = %+v, %v; want zero items, %v", got, err, tt.err)
			}
		})
	}
}

func TestStore_Create(t *testing.T) {
	t.Parallel()
	pool := testutil.Postgres(t)
	owner, err := user.NewStore(pool).Create(t.Context(), "Alice", "alice@example.com", "stored-password-hash")
	if err != nil {
		t.Fatal(err)
	}
	store := wallet.NewStore(pool)

	t.Run("concurrent creates commit with distinct sort orders", func(t *testing.T) {
		const count = 4
		type result struct {
			wallet wallet.Wallet
			err    error
		}
		results := make(chan result, count)
		start := make(chan struct{})
		for range count {
			go func() {
				<-start
				a, err := store.Create(t.Context(), owner.ID, &wallet.CreateWalletDTO{Name: "Savings", WalletType: string(wallet.Savings)})
				results <- result{a, err}
			}()
		}
		close(start)

		orders := make(map[int16]bool)
		for range count {
			got := <-results
			if got.err != nil {
				t.Errorf("Create() error = %v", got.err)
				continue
			}
			a := got.wallet
			if a.UserID != owner.ID || a.Name != "Savings" || a.Type != wallet.Savings || a.Balance != 0 || a.CreatedAt.IsZero() {
				t.Errorf("Create() = %+v, unexpected fields", a)
			}
			if a.SortOrder < 1 || a.SortOrder > count || orders[a.SortOrder] {
				t.Errorf("Create() sort order = %d, want distinct order in [1, %d]", a.SortOrder, count)
			}
			orders[a.SortOrder] = true
			persisted, err := store.GetByID(t.Context(), a.ID, owner.ID)
			if err != nil || !reflect.DeepEqual(persisted, a) {
				t.Errorf("persisted wallet = %+v, %v; want %+v", persisted, err, a)
			}
		}
		wallets, err := store.ListByUserID(t.Context(), owner.ID)
		if err != nil || len(wallets) != count+1 {
			t.Errorf("ListByUserID() count = %d, error = %v; want %d wallets including main", len(wallets), err, count+1)
		}
	})

	for _, tt := range []struct {
		name   string
		userID string
		err    error
	}{
		{name: "missing user", userID: "00000000-0000-0000-0000-000000000000", err: wallet.ErrNotFound},
		{name: "invalid user ID", userID: "invalid", err: wallet.ErrInvalidUserID},
		{name: "empty user ID", err: wallet.ErrInvalidUserID},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := store.Create(t.Context(), tt.userID, &wallet.CreateWalletDTO{Name: "Savings", WalletType: string(wallet.Savings)})
			if !errors.Is(err, tt.err) || !reflect.DeepEqual(got, wallet.Wallet{}) {
				t.Errorf("Create() = %+v, %v; want zero wallet, %v", got, err, tt.err)
			}
		})
	}
}
