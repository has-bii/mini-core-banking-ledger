package identifier

import "testing"

func TestIsValidEmail(t *testing.T) {
	for _, tc := range []struct {
		value string
		valid bool
	}{
		{"user@example.com", true},
		{"user.name+tag@example.co.uk", true},
		{"User Name <user@example.com>", true},
		{"", false},
		{"not-an-email", false},
		{"user@", false},
		{"@example.com", false},
		{"user name@example.com", false},
		{"user@@example.com", false},
		{"user@example.com, other@example.com", false},
		{"user@example.com\r\nBcc: other@example.com", false},
	} {
		t.Run(tc.value, func(t *testing.T) {
			if got := IsValidEmail(tc.value); got != tc.valid {
				t.Fatalf("isValidEmail(%q) = %v, want %v", tc.value, got, tc.valid)
			}
		})
	}
}
