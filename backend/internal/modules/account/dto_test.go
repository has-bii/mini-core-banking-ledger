package account_test

import (
	"backend/internal/modules/account"
	"backend/internal/platform/v2/validator"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestCreateAccountDTO_Validate(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name    string
		payload account.CreateAccountDTO
		want    validator.ErrorField
	}{
		{name: "valid and trimmed", payload: account.CreateAccountDTO{Name: " Savings Account ", AccountType: " savings "}},
		{name: "missing fields", want: validator.ErrorField{"name": {"This field is required"}, "account_type": {"This field is required"}}},
		{name: "first rule per field", payload: account.CreateAccountDTO{Name: "12", AccountType: "other"}, want: validator.ErrorField{"name": {"Must contain at least 3 characters"}, "account_type": {"Must be one of: sub_account savings"}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			payload := tt.payload
			assertAccountValidation(t, payload.Validate(), tt.want)
			if payload.Name != strings.TrimSpace(tt.payload.Name) || payload.AccountType != strings.TrimSpace(tt.payload.AccountType) {
				t.Fatalf("fields not normalized: %+v", payload)
			}
		})
	}
}

func TestCreateAccountDTO_NameValidation(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, value, message string }{
		{name: "blank", value: " \t", message: "This field is required"},
		{name: "two characters", value: "Ab", message: "Must contain at least 3 characters"},
		{name: "three characters", value: "Abc"},
		{name: "32 characters", value: strings.Repeat("A", 32)},
		{name: "33 characters", value: strings.Repeat("A", 33), message: "Must contain at most 32 characters"},
		{name: "digits", value: "Travel 123", message: `Failed validation rule "alphaspace"`},
		{name: "punctuation", value: "Travel!", message: `Failed validation rule "alphaspace"`},
		{name: "non ASCII", value: "Café", message: `Failed validation rule "alphaspace"`},
		{name: "internal tab", value: "My\tSavings", message: `Failed validation rule "alphaspace"`},
		{name: "internal newline", value: "My\nSavings", message: `Failed validation rule "alphaspace"`},
		{name: "internal spaces", value: "My  Savings"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			payload := account.CreateAccountDTO{Name: tt.value, AccountType: "savings"}
			var want validator.ErrorField
			if tt.message != "" {
				want = validator.ErrorField{"name": {tt.message}}
			}
			assertAccountValidation(t, payload.Validate(), want)
		})
	}
}

func TestCreateAccountDTO_TypeValidation(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, value, message string }{
		{name: "sub account", value: "sub_account"},
		{name: "savings", value: "savings"},
		{name: "main", value: "main", message: "Must be one of: sub_account savings"},
		{name: "missing", message: "This field is required"},
		{name: "whitespace", value: " \t", message: "This field is required"},
		{name: "unknown", value: "other", message: "Must be one of: sub_account savings"},
		{name: "uppercase", value: "SAVINGS", message: "Must be one of: sub_account savings"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			payload := account.CreateAccountDTO{Name: "Travel", AccountType: tt.value}
			var want validator.ErrorField
			if tt.message != "" {
				want = validator.ErrorField{"account_type": {tt.message}}
			}
			assertAccountValidation(t, payload.Validate(), want)
		})
	}
}

func assertAccountValidation(t *testing.T, err error, want validator.ErrorField) {
	t.Helper()
	if want == nil {
		if err != nil {
			t.Fatalf("Validate() = %v, want nil", err)
		}
		return
	}
	var got *validator.ErrorValidation
	if !errors.As(err, &got) || got.Errors == nil || !reflect.DeepEqual(*got.Errors, want) {
		t.Fatalf("Validate() = %v, want fields %+v", err, want)
	}
}
