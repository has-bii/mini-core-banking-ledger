package identifier_test

import (
	"testing"

	"backend/internal/platform/identifier"
)

func TestIsValidEmail(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		email string
		want  bool
	}{
		{"plain address", "alice@example.com", true},
		{"plus tag and subdomain", "alice+test@mail.example.com", true},
		{"display name", "Alice <alice@example.com>", false},
		{"angle brackets", "<alice@example.com>", false},
		{"surrounding whitespace", " alice@example.com ", false},
		{"empty", "", false},
		{"blank", " \t ", false},
		{"missing at sign", "alice.example.com", false},
		{"missing local part", "@example.com", false},
		{"missing domain", "alice@", false},
		{"multiple at signs", "alice@@example.com", false},
		{"multiple addresses", "alice@example.com, bob@example.com", false},
		{"newline", "alice@example.com\n", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := identifier.IsValidEmail(tt.email); got != tt.want {
				t.Errorf("IsValidEmail(%q) = %v; want %v", tt.email, got, tt.want)
			}
		})
	}
}
