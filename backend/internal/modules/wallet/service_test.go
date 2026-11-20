package wallet_test

import (
	"backend/internal/modules/wallet"
	"backend/internal/platform/validator"
	"context"
	"errors"
	"reflect"
	"testing"
)

type walletStoreStub struct {
	wallets []wallet.Wallet
	wallet  wallet.Wallet
	id      string
	err     error
	ctx     context.Context
	userID  string
	data    *wallet.CreateWalletDTO
}

func (s *walletStoreStub) ListByUserID(ctx context.Context, userID string) ([]wallet.Wallet, error) {
	s.ctx, s.userID = ctx, userID
	return s.wallets, s.err
}

func (s *walletStoreStub) GetByID(ctx context.Context, id, userID string) (wallet.Wallet, error) {
	s.ctx, s.id, s.userID = ctx, id, userID
	return s.wallet, s.err
}

func (s *walletStoreStub) Create(ctx context.Context, userID string, data *wallet.CreateWalletDTO) (wallet.Wallet, error) {
	s.ctx, s.userID, s.data = ctx, userID, data
	return s.wallet, s.err
}

func TestService_GetUserWalletPreservesCause(t *testing.T) {
	t.Parallel()
	cause := errors.New("lookup failed")
	store := &walletStoreStub{err: errors.Join(cause, wallet.ErrNotFound)}
	_, err := wallet.NewService(store).GetUserWallet(t.Context(), "id", "user")
	if !errors.Is(err, wallet.ErrNotFound) || !errors.Is(err, cause) {
		t.Fatalf("GetUserWallet() error = %v; want ErrNotFound preserving cause", err)
	}
}

func TestService_GetUserWallet(t *testing.T) {
	t.Parallel()
	const userID = "01a1c6de-805e-7b0c-a1b1-b851a9c38ccd"
	const id = "01a1c6de-805e-7b0c-a1b1-b851a9c38cce"

	for _, tt := range []struct {
		name   string
		id     string
		userID string
		wallet wallet.Wallet
		err    error
	}{
		{name: "returns wallet", id: id, userID: userID, wallet: wallet.Wallet{ID: id, UserID: userID, Name: "Main Wallet", Type: wallet.Main, Balance: 12345}},
		{name: "not found", id: id, userID: userID, err: wallet.ErrNotFound},
		{name: "invalid wallet ID", id: "invalid", userID: userID, err: wallet.ErrInvalidWalletID},
		{name: "invalid user ID", id: id, userID: "invalid", err: wallet.ErrInvalidUserID},
		{name: "store error", id: id, userID: userID, err: errors.New("store unavailable")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := &walletStoreStub{wallet: tt.wallet, err: tt.err}
			ctx := t.Context()

			got, err := wallet.NewService(store).GetUserWallet(ctx, tt.id, tt.userID)
			if !errors.Is(err, tt.err) {
				t.Errorf("GetUserWallet() error = %v, want %v", err, tt.err)
			}
			if !reflect.DeepEqual(got, tt.wallet) {
				t.Errorf("GetUserWallet() = %+v, want %+v", got, tt.wallet)
			}
			if store.id != tt.id || store.userID != tt.userID || store.ctx != ctx {
				t.Errorf("GetByID() did not receive original arguments: id = %q, userID = %q, ctx = %v", store.id, store.userID, store.ctx)
			}
		})
	}
}

func TestService_GetUserWallets(t *testing.T) {
	t.Parallel()
	const userID = "01a1c6de-805e-7b0c-a1b1-b851a9c38ccd"
	storeErr := errors.New("store unavailable")

	for _, tt := range []struct {
		name    string
		userID  string
		wallets []wallet.Wallet
		err     error
	}{
		{
			name:   "returns wallets",
			userID: userID,
			wallets: []wallet.Wallet{
				{ID: "01a1c6de-805e-7b0c-a1b1-b851a9c38cce", Name: "Main Wallet", Type: wallet.Main, Balance: 12345},
			},
		},
		{
			name:    "no wallets",
			userID:  userID,
			wallets: []wallet.Wallet{},
		},
		{
			name: "invalid user ID",
			err:  wallet.ErrInvalidUserID,
		},
		{
			name:   "store error",
			userID: userID,
			err:    storeErr,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := &walletStoreStub{wallets: tt.wallets, err: tt.err}
			service := wallet.NewService(store)
			ctx := t.Context()

			got, err := service.GetUserWallets(ctx, tt.userID)
			if !errors.Is(err, tt.err) {
				t.Errorf("GetUserWallets() error = %v, want %v", err, tt.err)
			}
			if !reflect.DeepEqual(got, tt.wallets) {
				t.Errorf("GetUserWallets() = %+v, want %+v", got, tt.wallets)
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

func TestService_CreateUserWalletPreservesErrors(t *testing.T) {
	t.Parallel()
	fields := validator.ErrorField{"name": {"This field is required"}}
	validationErr := &validator.ErrorValidation{Errors: &fields}
	for _, tt := range []struct {
		name  string
		cause error
	}{
		{name: "missing owner", cause: wallet.ErrNotFound},
		{name: "invalid user ID", cause: wallet.ErrInvalidUserID},
		{name: "invalid fields", cause: validationErr},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := wallet.NewService(&walletStoreStub{err: tt.cause}).CreateUserWallet(t.Context(), "user", &wallet.CreateWalletDTO{})
			if !errors.Is(err, tt.cause) {
				t.Fatalf("CreateUserWallet() error = %v; want original cause %v", err, tt.cause)
			}
			if tt.cause == validationErr {
				var got *validator.ErrorValidation
				if !errors.As(err, &got) || !reflect.DeepEqual(*got.Errors, fields) {
					t.Fatalf("validation details lost: %v", err)
				}
			}
		})
	}
}

func TestService_CreateUserWallet(t *testing.T) {
	t.Parallel()
	const userID = "01a1c6de-805e-7b0c-a1b1-b851a9c38ccd"
	storeErr := errors.New("store unavailable")
	store := &walletStoreStub{err: storeErr}
	ctx := t.Context()
	data := &wallet.CreateWalletDTO{Name: "Travel", WalletType: string(wallet.Savings)}
	got, err := wallet.NewService(store).CreateUserWallet(ctx, userID, data)
	if !errors.Is(err, storeErr) || got != (wallet.Wallet{}) {
		t.Errorf("CreateUserWallet() = %+v, %v; want zero wallet, %v", got, err, storeErr)
	}
	if store.ctx != ctx || store.userID != userID || store.data != data {
		t.Error("Create() did not receive original context, authenticated user ID and payload")
	}
}
