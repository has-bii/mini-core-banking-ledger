package account_test

import (
	"backend/internal/modules/account"
	"backend/internal/platform/v2/validator"
	"errors"
	"reflect"
	"testing"
)

func TestStore_RejectsInvalidIDsBeforeSQL(t *testing.T) {
	t.Parallel()
	const validID = "01a1c6de-805e-7b0c-a1b1-b851a9c38ccd"
	for _, tt := range []struct {
		name string
		call func(*account.Store) error
		want error
	}{
		{name: "list invalid user", call: func(s *account.Store) error { _, err := s.ListByUserID(t.Context(), "invalid"); return err }, want: account.ErrInvalidUserID},
		{name: "get invalid account", call: func(s *account.Store) error { _, err := s.GetByID(t.Context(), "invalid", validID); return err }, want: account.ErrInvalidAccountID},
		{name: "get empty account", call: func(s *account.Store) error { _, err := s.GetByID(t.Context(), "", validID); return err }, want: account.ErrInvalidAccountID},
		{name: "get invalid user", call: func(s *account.Store) error { _, err := s.GetByID(t.Context(), validID, "invalid"); return err }, want: account.ErrInvalidUserID},
		{name: "create invalid user", call: func(s *account.Store) error {
			_, err := s.Create(t.Context(), "invalid", &account.CreateAccountDTO{})
			return err
		}, want: account.ErrInvalidUserID},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if err := tt.call(account.NewStore(nil)); !errors.Is(err, tt.want) {
				t.Fatalf("error = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestStore_CreatePreservesValidationErrors(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name    string
		payload account.CreateAccountDTO
		want    validator.ErrorField
	}{
		{name: "invalid name", payload: account.CreateAccountDTO{Name: "12", AccountType: string(account.Savings)}, want: validator.ErrorField{
			"name": {"Must contain at least 3 characters"},
		}},
		{name: "invalid type", payload: account.CreateAccountDTO{Name: "Travel", AccountType: string(account.Main)}, want: validator.ErrorField{
			"account_type": {"Must be one of: sub_account savings"},
		}},
		{name: "multiple fields", payload: account.CreateAccountDTO{}, want: validator.ErrorField{
			"name":         {"This field is required"},
			"account_type": {"This field is required"},
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Nil storage proves invalid payloads are rejected before any SQL.
			got, err := account.NewStore(nil).Create(t.Context(), "01a1c6de-805e-7b0c-a1b1-b851a9c38ccd", &tt.payload)
			var fields *validator.ErrorValidation
			if !errors.As(err, &fields) || !reflect.DeepEqual(*fields.Errors, tt.want) || got != (account.Account{}) {
				t.Errorf("Create() = %+v, %v; want zero account and validation errors %+v", got, err, tt.want)
			}
		})
	}
}
