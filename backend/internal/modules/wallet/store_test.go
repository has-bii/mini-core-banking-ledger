package wallet_test

import (
	"backend/internal/modules/wallet"
	"backend/internal/platform/validator"
	"errors"
	"reflect"
	"testing"
)

func TestStore_RejectsInvalidIDsBeforeSQL(t *testing.T) {
	t.Parallel()
	const validID = "01a1c6de-805e-7b0c-a1b1-b851a9c38ccd"
	for _, tt := range []struct {
		name string
		call func(*wallet.Store) error
		want error
	}{
		{name: "list invalid user", call: func(s *wallet.Store) error { _, err := s.ListByUserID(t.Context(), "invalid"); return err }, want: wallet.ErrInvalidUserID},
		{name: "get invalid wallet", call: func(s *wallet.Store) error { _, err := s.GetByID(t.Context(), "invalid", validID); return err }, want: wallet.ErrInvalidWalletID},
		{name: "get empty wallet", call: func(s *wallet.Store) error { _, err := s.GetByID(t.Context(), "", validID); return err }, want: wallet.ErrInvalidWalletID},
		{name: "get invalid user", call: func(s *wallet.Store) error { _, err := s.GetByID(t.Context(), validID, "invalid"); return err }, want: wallet.ErrInvalidUserID},
		{name: "create invalid user", call: func(s *wallet.Store) error {
			_, err := s.Create(t.Context(), "invalid", &wallet.CreateWalletDTO{})
			return err
		}, want: wallet.ErrInvalidUserID},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := tt.call(wallet.NewStore(nil)); !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestStore_CreatePreservesValidationErrors(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name    string
		payload wallet.CreateWalletDTO
		want    validator.ErrorField
	}{
		{name: "invalid name", payload: wallet.CreateWalletDTO{Name: "12", WalletType: string(wallet.Savings)}, want: validator.ErrorField{
			"name": {"Must contain at least 3 characters"},
		}},
		{name: "invalid type", payload: wallet.CreateWalletDTO{Name: "Travel", WalletType: string(wallet.Main)}, want: validator.ErrorField{
			"wallet_type": {"Must be one of: sub_wallet savings"},
		}},
		{name: "multiple fields", payload: wallet.CreateWalletDTO{}, want: validator.ErrorField{
			"name":        {"This field is required"},
			"wallet_type": {"This field is required"},
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Nil storage proves invalid payloads are rejected before any SQL.
			got, err := wallet.NewStore(nil).Create(t.Context(), "01a1c6de-805e-7b0c-a1b1-b851a9c38ccd", &tt.payload)
			var fields *validator.ErrorValidation
			if !errors.As(err, &fields) || !reflect.DeepEqual(*fields.Errors, tt.want) || got != (wallet.Wallet{}) {
				t.Errorf("Create() = %+v, %v; want zero wallet and validation errors %+v", got, err, tt.want)
			}
		})
	}
}
