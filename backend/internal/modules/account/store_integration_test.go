//go:build integration

package account_test

import (
	"backend/internal/modules/account"
	"backend/internal/modules/user"
	"backend/internal/platform/validation"
	"backend/internal/testutil"
	"errors"
	"testing"
	"time"
)

func TestStore_FindAllByUserID(t *testing.T) {
	t.Parallel()
	db := testutil.Postgres(t)
	users := user.NewStore(db)
	alice, err := users.Create(t.Context(), "Alice", "alice@example.com", "stored-password-hash")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := users.Create(t.Context(), "Bob", "bob@example.com", "stored-password-hash"); err != nil {
		t.Fatal(err)
	}
	// Insert out of order so the assertion exercises SQL ordering.
	if _, err := db.Exec(t.Context(), `INSERT INTO accounts (id, user_id, name, type, sort_order, balance, created_at)
		VALUES ('00000000-0000-0000-0000-000000000002', $1, 'Savings', 'savings', 2, 12345, '2026-01-02T03:04:05Z'),
		       ('00000000-0000-0000-0000-000000000001', $1, 'Spending', 'sub_account', 1, 678, '2026-01-02T03:04:05Z')`, alice.ID); err != nil {
		t.Fatal(err)
	}

	store := account.NewStore(db)
	got, err := store.FindAllByUserID(t.Context(), alice.ID, validation.ASC)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("accounts = %+v, want Alice's three accounts only", got)
	}
	if got[0].ID == "" || got[0].Name != "Main Account" || got[0].Type != account.Main || got[0].SortOrder != 0 || got[0].Balance != 0 || got[0].CreatedAt.IsZero() {
		t.Errorf("first account = %+v, want populated main account", got[0])
	}
	createdAt := time.Date(2026, time.January, 2, 3, 4, 5, 0, time.UTC)
	want := []account.Account{
		{ID: "00000000-0000-0000-0000-000000000001", Name: "Spending", Type: account.SubAccount, SortOrder: 1, Balance: 678, CreatedAt: createdAt},
		{ID: "00000000-0000-0000-0000-000000000002", Name: "Savings", Type: account.Savings, SortOrder: 2, Balance: 12345, CreatedAt: createdAt},
	}
	for i, expected := range want {
		actual := got[i+1]
		if actual.ID != expected.ID || actual.Name != expected.Name || actual.Type != expected.Type || actual.SortOrder != expected.SortOrder || actual.Balance != expected.Balance || !actual.CreatedAt.Equal(expected.CreatedAt) {
			t.Errorf("account[%d] = %+v, want %+v", i+1, actual, expected)
		}
	}
	descending, err := store.FindAllByUserID(t.Context(), alice.ID, validation.DESC)
	if err != nil {
		t.Fatal(err)
	}
	if len(descending) != 3 || descending[0].ID != want[1].ID || descending[1].ID != want[0].ID || descending[2].ID != got[0].ID {
		t.Errorf("descending accounts = %+v, want Savings, Spending, Main Account", descending)
	}
	if _, err := store.FindAllByUserID(t.Context(), alice.ID, validation.QueryOrder("desc; DROP TABLE accounts")); !errors.Is(err, validation.ErrValidationError) {
		t.Errorf("invalid order error = %v, want ErrValidationError", err)
	}
	missing, err := store.FindAllByUserID(t.Context(), "00000000-0000-0000-0000-000000000000", validation.ASC)
	if err != nil || len(missing) != 0 {
		t.Errorf("unknown user's accounts = %+v, %v; want empty result, nil", missing, err)
	}
}
