package identifier_test

import (
	"testing"

	"backend/internal/platform/identifier"
)

func TestValidUUID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"canonical", "550e8400-e29b-41d4-a716-446655440000", true},
		{"uppercase", "550E8400-E29B-41D4-A716-446655440000", true},
		{"nil UUID", "00000000-0000-0000-0000-000000000000", true},
		{"without hyphens", "550e8400e29b41d4a716446655440000", true},
		{"URN", "urn:uuid:550e8400-e29b-41d4-a716-446655440000", true},
		{"braces", "{550e8400-e29b-41d4-a716-446655440000}", true},
		{"empty", "", false},
		{"blank", " \t ", false},
		{"invalid length", "550e8400-e29b-41d4-a716-44665544000", false},
		{"invalid hex", "550e8400-e29b-41d4-a716-44665544000g", false},
		{"invalid separator", "550e8400_e29b-41d4-a716-446655440000", false},
		{"invalid URN prefix", "bad:uuid:550e8400-e29b-41d4-a716-446655440000", false},
		{"invalid braces", "[550e8400-e29b-41d4-a716-446655440000]", false},
		{"leading whitespace", " 550e8400-e29b-41d4-a716-446655440000", false},
		{"trailing whitespace", "550e8400-e29b-41d4-a716-446655440000 ", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := identifier.ValidUUID(tt.value); got != tt.want {
				t.Errorf("ValidUUID(%q) = %v; want %v", tt.value, got, tt.want)
			}
		})
	}
}
