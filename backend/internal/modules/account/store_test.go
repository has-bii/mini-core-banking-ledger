package account_test

import (
	"backend/internal/modules/account"
	"backend/internal/platform/validation"
	"errors"
	"reflect"
	"testing"
)

func TestStore_CreatePreservesValidationErrors(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name    string
		payload account.CreateAccountDTO
		want    validation.Errors
	}{
		{"invalid name", account.CreateAccountDTO{Name: "12", Type: account.Savings}, validation.Errors{
			{Field: "name", Code: "MIN_LENGTH", Message: "Name must contain at least 3 characters."},
		}},
		{"invalid type", account.CreateAccountDTO{Name: "Travel", Type: account.Main}, validation.Errors{
			{Field: "account_type", Code: "INVALID_CHOICE", Message: `Account type must be "sub_account" or "savings".`},
		}},
		{"multiple fields", account.CreateAccountDTO{}, validation.Errors{
			{Field: "name", Code: "REQUIRED", Message: "Name is required."},
			{Field: "account_type", Code: "REQUIRED", Message: "Account type is required."},
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Nil storage proves invalid payloads are rejected before any SQL.
			got, err := account.NewStore(nil).Create(t.Context(), "01a1c6de-805e-7b0c-a1b1-b851a9c38ccd", &tt.payload)
			var fields validation.Errors
			if !errors.As(err, &fields) || !reflect.DeepEqual(fields, tt.want) || got != (account.Account{}) {
				t.Errorf("Create() = %+v, %v; want zero account and validation errors %+v", got, err, tt.want)
			}
		})
	}
}
