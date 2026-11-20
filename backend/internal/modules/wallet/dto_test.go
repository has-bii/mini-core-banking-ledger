package wallet_test

import (
	"backend/internal/modules/wallet"
	"backend/internal/platform/validator"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func TestCreateWalletDTO_Validate(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name    string
		payload wallet.CreateWalletDTO
		want    validator.ErrorField
	}{
		{name: "valid and trimmed", payload: wallet.CreateWalletDTO{Name: " Savings Wallet ", WalletType: " savings "}},
		{name: "missing fields", want: validator.ErrorField{"name": {"This field is required"}, "wallet_type": {"This field is required"}}},
		{name: "first rule per field", payload: wallet.CreateWalletDTO{Name: "12", WalletType: "other"}, want: validator.ErrorField{"name": {"Must contain at least 3 characters"}, "wallet_type": {"Must be one of: sub_wallet savings"}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			payload := tt.payload
			assertWalletValidation(t, payload.Validate(), tt.want)
			if payload.Name != strings.TrimSpace(tt.payload.Name) || payload.WalletType != strings.TrimSpace(tt.payload.WalletType) {
				t.Fatalf("fields not normalized: %+v", payload)
			}
		})
	}
}

func TestCreateWalletDTO_NameValidation(t *testing.T) {
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
			payload := wallet.CreateWalletDTO{Name: tt.value, WalletType: "savings"}
			var want validator.ErrorField
			if tt.message != "" {
				want = validator.ErrorField{"name": {tt.message}}
			}
			assertWalletValidation(t, payload.Validate(), want)
		})
	}
}

func TestCreateWalletDTO_TypeValidation(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, value, message string }{
		{name: "sub wallet", value: "sub_wallet"},
		{name: "savings", value: "savings"},
		{name: "main", value: "main", message: "Must be one of: sub_wallet savings"},
		{name: "missing", message: "This field is required"},
		{name: "whitespace", value: " \t", message: "This field is required"},
		{name: "unknown", value: "other", message: "Must be one of: sub_wallet savings"},
		{name: "uppercase", value: "SAVINGS", message: "Must be one of: sub_wallet savings"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			payload := wallet.CreateWalletDTO{Name: "Travel", WalletType: tt.value}
			var want validator.ErrorField
			if tt.message != "" {
				want = validator.ErrorField{"wallet_type": {tt.message}}
			}
			assertWalletValidation(t, payload.Validate(), want)
		})
	}
}

func assertWalletValidation(t *testing.T, err error, want validator.ErrorField) {
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
