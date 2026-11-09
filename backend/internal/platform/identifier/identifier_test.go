package identifier_test

import (
	"testing"

	"backend/internal/platform/identifier"
)

func TestIsValidEmail(t *testing.T) {
	for _, tt := range []struct {
		value string
		want  bool
	}{
		{"alice@example.com", true},
		{"alice+bank@example.co.id", true},
		{"", false},
		{"alice", false},
		{"alice@", false},
		{"Alice <alice@example.com>", false},
		{" alice@example.com ", false},
	} {
		t.Run(tt.value, func(t *testing.T) {
			if got := identifier.IsValidEmail(tt.value); got != tt.want {
				t.Errorf("IsValidEmail(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}

func TestValidUUID(t *testing.T) {
	for _, tt := range []struct {
		value string
		want  bool
	}{
		{"550e8400-e29b-41d4-a716-446655440000", true},
		{"0197b4b4-25e3-7000-8000-000000000001", true},
		{"", false},
		{"not-a-uuid", false},
		{"550e8400-e29b-41d4-a716-44665544000g", false},
		{"550e8400-e29b-41d4-a716-44665544000", false},
	} {
		t.Run(tt.value, func(t *testing.T) {
			if got := identifier.ValidUUID(tt.value); got != tt.want {
				t.Errorf("ValidUUID(%q) = %v, want %v", tt.value, got, tt.want)
			}
		})
	}
}
