package account_test

import (
	"backend/internal/modules/account"
	"backend/internal/platform/validation"
	"slices"
	"strings"
	"testing"
)

func TestCreateAccountDTO_Validate(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name    string
		payload account.CreateAccountDTO
		want    validation.Errors
	}{
		{"valid and trimmed", account.CreateAccountDTO{Name: " Savings Account ", Type: " savings "}, nil},
		{"missing fields", account.CreateAccountDTO{}, validation.Errors{
			{Field: "name", Code: "REQUIRED", Message: "Name is required."},
			{Field: "account_type", Code: "REQUIRED", Message: "Account type is required."},
		}},
		{"multiple errors with first rule per field", account.CreateAccountDTO{Name: "12", Type: "other"}, validation.Errors{
			{Field: "name", Code: "MIN_LENGTH", Message: "Name must contain at least 3 characters."},
			{Field: "account_type", Code: "INVALID_CHOICE", Message: `Account type must be "sub_account" or "savings".`},
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			payload := tt.payload
			if got := payload.Validate(); !slices.Equal(got, tt.want) {
				t.Errorf("Validate() = %+v, want %+v", got, tt.want)
			}
			if payload.Name != strings.TrimSpace(tt.payload.Name) || payload.Type != account.AccountType(strings.TrimSpace(string(tt.payload.Type))) {
				t.Errorf("normalization = %+v, want trimmed name and type", payload)
			}
		})
	}
}

func TestCreateAccountDTO_NameValidation(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, value, code, message string }{
		{"blank", " \t", "REQUIRED", "Name is required."},
		{"two characters", "Ab", "MIN_LENGTH", "Name must contain at least 3 characters."},
		{"three characters", "Abc", "", ""},
		{"32 characters", strings.Repeat("A", 32), "", ""},
		{"33 characters", strings.Repeat("A", 33), "MAX_LENGTH", "Name must contain at most 32 characters."},
		{"digits", "Travel 123", "INVALID_FORMAT", "Name must contain only ASCII letters or spaces."},
		{"punctuation", "Travel!", "INVALID_FORMAT", "Name must contain only ASCII letters or spaces."},
		{"non ASCII", "Café", "INVALID_FORMAT", "Name must contain only ASCII letters or spaces."},
		{"internal tab", "My\tSavings", "INVALID_FORMAT", "Name must contain only ASCII letters or spaces."},
		{"internal newline", "My\nSavings", "INVALID_FORMAT", "Name must contain only ASCII letters or spaces."},
		{"internal spaces", "My  Savings", "", ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			payload := account.CreateAccountDTO{Name: tt.value, Type: account.Savings}
			var want validation.Errors
			if tt.code != "" {
				want = validation.Errors{{Field: "name", Code: tt.code, Message: tt.message}}
			}
			if got := payload.Validate(); !slices.Equal(got, want) {
				t.Errorf("Validate() = %+v, want %+v", got, want)
			}
		})
	}
}

func TestCreateAccountDTO_TypeValidation(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, code, message string
		value               account.AccountType
	}{
		{"sub account", "", "", account.SubAccount},
		{"savings", "", "", account.Savings},
		{"main", "INVALID_CHOICE", `Account type must be "sub_account" or "savings".`, account.Main},
		{"missing", "REQUIRED", "Account type is required.", ""},
		{"whitespace", "REQUIRED", "Account type is required.", " \t"},
		{"unknown", "INVALID_CHOICE", `Account type must be "sub_account" or "savings".`, "other"},
		{"uppercase", "INVALID_CHOICE", `Account type must be "sub_account" or "savings".`, "SAVINGS"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			payload := account.CreateAccountDTO{Name: "Travel", Type: tt.value}
			var want validation.Errors
			if tt.code != "" {
				want = validation.Errors{{Field: "account_type", Code: tt.code, Message: tt.message}}
			}
			if got := payload.Validate(); !slices.Equal(got, want) {
				t.Errorf("Validate() = %+v, want %+v", got, want)
			}
		})
	}
}
