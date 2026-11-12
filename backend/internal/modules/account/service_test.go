package account_test

import (
	"backend/internal/modules/account"
	"backend/internal/platform/apperror"
	"backend/internal/platform/validation"
	"context"
	"errors"
	"reflect"
	"testing"
)

type accountStoreStub struct {
	accounts []account.Account
	account  account.Account
	id       string
	err      error
	ctx      context.Context
	userID   string
	data     *account.CreateAccountDTO
}

func (s *accountStoreStub) ListByUserID(ctx context.Context, userID string) ([]account.Account, error) {
	s.ctx, s.userID = ctx, userID
	return s.accounts, s.err
}

func (s *accountStoreStub) GetByID(ctx context.Context, id, userID string) (account.Account, error) {
	s.ctx, s.id, s.userID = ctx, id, userID
	return s.account, s.err
}

func (s *accountStoreStub) Create(ctx context.Context, userID string, data *account.CreateAccountDTO) (account.Account, error) {
	s.ctx, s.userID, s.data = ctx, userID, data
	return s.account, s.err
}

func TestService_GetUserAccountApplicationError(t *testing.T) {
	t.Parallel()
	cause := errors.New("lookup failed")
	store := &accountStoreStub{err: errors.Join(cause, account.ErrNotFound)}
	_, err := account.NewService(store).GetUserAccount(t.Context(), "id", "user")
	var appErr *apperror.Error
	if !errors.As(err, &appErr) || appErr.Code != "NOT_FOUND" || appErr.Message != "account not found" || !errors.Is(err, cause) {
		t.Fatalf("GetUserAccount() error = %v; want NOT_FOUND preserving cause", err)
	}
}

func TestService_GetUserAccount(t *testing.T) {
	t.Parallel()
	const userID = "01a1c6de-805e-7b0c-a1b1-b851a9c38ccd"
	const id = "01a1c6de-805e-7b0c-a1b1-b851a9c38cce"

	for _, tt := range []struct {
		name    string
		id      string
		userID  string
		account account.Account
		err     error
	}{
		{name: "returns account", id: id, userID: userID, account: account.Account{ID: id, UserID: userID, Name: "Main Account", Type: account.Main, Balance: 12345}},
		{name: "not found", id: id, userID: userID, err: account.ErrNotFound},
		{name: "invalid account ID", id: "invalid", userID: userID, err: validation.ErrInvalidUUID},
		{name: "invalid user ID", id: id, userID: "invalid", err: validation.ErrInvalidUUID},
		{name: "store error", id: id, userID: userID, err: errors.New("store unavailable")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := &accountStoreStub{account: tt.account, err: tt.err}
			ctx := t.Context()

			got, err := account.NewService(store).GetUserAccount(ctx, tt.id, tt.userID)
			if !errors.Is(err, tt.err) {
				t.Errorf("GetUserAccount() error = %v, want %v", err, tt.err)
			}
			if !reflect.DeepEqual(got, tt.account) {
				t.Errorf("GetUserAccount() = %+v, want %+v", got, tt.account)
			}
			if store.id != tt.id || store.userID != tt.userID || store.ctx != ctx {
				t.Errorf("GetByID() did not receive original arguments: id = %q, userID = %q, ctx = %v", store.id, store.userID, store.ctx)
			}
		})
	}
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

func TestService_CreateUserAccountApplicationErrors(t *testing.T) {
	t.Parallel()
	fields := validation.Errors{{Field: "name", Code: "REQUIRED", Message: "Name is required."}}
	for _, tt := range []struct {
		name    string
		cause   error
		code    string
		details any
	}{
		{name: "missing owner", cause: account.ErrNotFound, code: "NOT_FOUND"},
		{name: "invalid fields", cause: fields, code: "VALIDATION_ERROR", details: fields},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := account.NewService(&accountStoreStub{err: tt.cause}).CreateUserAccount(t.Context(), "user", &account.CreateAccountDTO{})
			var appErr *apperror.Error
			if !errors.As(err, &appErr) || appErr.Code != tt.code || !reflect.DeepEqual(appErr.Details, tt.details) || appErr.Unwrap() == nil {
				t.Fatalf("CreateUserAccount() error = %v; want %s with details %+v and cause", err, tt.code, tt.details)
			}
		})
	}
}

func TestService_CreateUserAccount(t *testing.T) {
	t.Parallel()
	const userID = "01a1c6de-805e-7b0c-a1b1-b851a9c38ccd"
	storeErr := errors.New("store unavailable")
	store := &accountStoreStub{err: storeErr}
	ctx := t.Context()
	data := &account.CreateAccountDTO{Name: "Travel", Type: account.Savings}
	got, err := account.NewService(store).CreateUserAccount(ctx, userID, data)
	if !errors.Is(err, storeErr) || got != (account.Account{}) {
		t.Errorf("CreateUserAccount() = %+v, %v; want zero account, %v", got, err, storeErr)
	}
	if store.ctx != ctx || store.userID != userID || store.data != data {
		t.Error("Create() did not receive original context, authenticated user ID and payload")
	}
}
