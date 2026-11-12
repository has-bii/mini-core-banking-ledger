package account_test

import (
	"backend/internal/modules/account"
	"slices"
	"strings"
	"testing"
)

func TestCreateAccountDTO_Validate(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name    string
		payload account.CreateAccountDTO
		want    []error
	}{
		{"valid and trimmed", account.CreateAccountDTO{Name: " Savings Account ", Type: " savings "}, nil},
		{"main", account.CreateAccountDTO{Name: "Main", Type: account.Main}, []error{account.ErrInvalidType}},
		{"sub account", account.CreateAccountDTO{Name: "Travel", Type: account.SubAccount}, nil},
		{"missing fields", account.CreateAccountDTO{}, []error{account.ErrNameRequired, account.ErrInvalidType}},
		{"blank name", account.CreateAccountDTO{Name: " \t", Type: account.Savings}, []error{account.ErrNameRequired}},
		{"two characters", account.CreateAccountDTO{Name: "Ab", Type: account.Savings}, []error{account.ErrInvalidName}},
		{"three characters", account.CreateAccountDTO{Name: "Abc", Type: account.Savings}, nil},
		{"32 characters", account.CreateAccountDTO{Name: strings.Repeat("A", 32), Type: account.Savings}, nil},
		{"33 characters", account.CreateAccountDTO{Name: strings.Repeat("A", 33), Type: account.Savings}, []error{account.ErrInvalidName}},
		{"digits", account.CreateAccountDTO{Name: "Travel 123", Type: account.Savings}, []error{account.ErrInvalidName}},
		{"punctuation", account.CreateAccountDTO{Name: "Travel!", Type: account.Savings}, []error{account.ErrInvalidName}},
		{"non ASCII", account.CreateAccountDTO{Name: "Café", Type: account.Savings}, []error{account.ErrInvalidName}},
		{"internal tab", account.CreateAccountDTO{Name: "My\tSavings", Type: account.Savings}, []error{account.ErrInvalidName}},
		{"internal newline", account.CreateAccountDTO{Name: "My\nSavings", Type: account.Savings}, []error{account.ErrInvalidName}},
		{"internal spaces", account.CreateAccountDTO{Name: "My  Savings", Type: account.Savings}, nil},
		{"missing type", account.CreateAccountDTO{Name: "Travel"}, []error{account.ErrInvalidType}},
		{"unknown type", account.CreateAccountDTO{Name: "Travel", Type: "other"}, []error{account.ErrInvalidType}},
		{"uppercase type", account.CreateAccountDTO{Name: "Travel", Type: "SAVINGS"}, []error{account.ErrInvalidType}},
		{"multiple errors", account.CreateAccountDTO{Name: "12", Type: "other"}, []error{account.ErrInvalidName, account.ErrInvalidType}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			payload := tt.payload
			if got := payload.Validate(); !slices.Equal(got, tt.want) {
				t.Fatalf("Validate() = %v, want %v", got, tt.want)
			}
			if payload.Name != strings.TrimSpace(tt.payload.Name) || payload.Type != account.AccountType(strings.TrimSpace(string(tt.payload.Type))) {
				t.Errorf("normalization = %+v, want trimmed name and type", payload)
			}
		})
	}
}
