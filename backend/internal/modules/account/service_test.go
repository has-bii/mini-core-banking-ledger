package account_test

import (
	"backend/internal/modules/account"
	"backend/internal/platform/validation"
	"context"
	"errors"
	"reflect"
	"testing"
)

type accountStoreStub struct {
	accounts []account.Account
	err      error
	ctx      context.Context
	userID   string
}

func (s *accountStoreStub) ListByUserID(ctx context.Context, userID string) ([]account.Account, error) {
	s.ctx, s.userID = ctx, userID
	return s.accounts, s.err
}

func TestService_GetUserAccounts(t *testing.T) {
	t.Parallel()
	const userID = "01a1c6de-805e-7b0c-a1b1-b851a9c38ccd"
	storeErr := errors.New("store unavailable")

	for _, tt := range []struct {
		name     string
		userID   string
		accounts []account.Account
		err      error
	}{
		{
			name:   "returns accounts",
			userID: userID,
			accounts: []account.Account{
				{ID: "01a1c6de-805e-7b0c-a1b1-b851a9c38cce", Name: "Main Account", Type: account.Main, Balance: 12345},
			},
		},
		{
			name:     "no accounts",
			userID:   userID,
			accounts: []account.Account{},
		},
		{
			name: "invalid user ID",
			err:  validation.ErrInvalidUUID,
		},
		{
			name:   "store error",
			userID: userID,
			err:    storeErr,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := &accountStoreStub{accounts: tt.accounts, err: tt.err}
			service := account.NewService(store)
			ctx := t.Context()

			got, err := service.GetUserAccounts(ctx, tt.userID)
			if !errors.Is(err, tt.err) {
				t.Errorf("GetUserAccounts() error = %v, want %v", err, tt.err)
			}
			if !reflect.DeepEqual(got, tt.accounts) {
				t.Errorf("GetUserAccounts() = %+v, want %+v", got, tt.accounts)
			}
			if store.userID != tt.userID {
				t.Errorf("ListByUserID() userID = %q, want %q", store.userID, tt.userID)
			}
			if store.ctx != ctx {
				t.Error("ListByUserID() did not receive the original context")
			}
		})
	}
}
