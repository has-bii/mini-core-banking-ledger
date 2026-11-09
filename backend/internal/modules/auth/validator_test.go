package auth

import (
	"strings"
	"testing"
)

func TestRegisterPayloadValidateCollectsErrors(t *testing.T) {
	parsed, errs := (RegisterPayload{}).Validate()
	if parsed != (RegisterPayload{}) {
		t.Fatalf("invalid payload returned: %+v", parsed)
	}
	want := []string{
		"name is required",
		"email must be a valid email address",
		"password must contain at least 8 characters",
	}
	if len(errs) != len(want) {
		t.Fatalf("Validate() = %v, want %d errors", errs, len(want))
	}
	for i, message := range want {
		if errs[i].Error() != message {
			t.Errorf("error %d = %q, want %q", i, errs[i].Error(), message)
		}
	}
}

func TestRegisterPayloadValidateNormalizes(t *testing.T) {
	original := RegisterPayload{" Alice \t", " alice@example.com ", " password123 "}
	parsed, errs := original.Validate()
	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}
	want := RegisterPayload{"Alice", "alice@example.com", " password123 "}
	if parsed != want {
		t.Fatal("expected trimmed name and email with unchanged password")
	}
	if original.Name != " Alice \t" || original.Email != " alice@example.com " {
		t.Fatal("validation modified the original payload")
	}
}

func TestRegisterPayloadValidate(t *testing.T) {
	tests := []struct {
		name      string
		payload   RegisterPayload
		wantError string
	}{
		{"valid", RegisterPayload{"Alice", "alice@example.com", "password123"}, ""},
		{"blank name", RegisterPayload{" \t", "alice@example.com", "password123"}, "name is required"},
		{"empty email", RegisterPayload{"Alice", "", "password123"}, "email must be a valid email address"},
		{"invalid email", RegisterPayload{"Alice", "not-an-email", "password123"}, "email must be a valid email address"},
		{"display name", RegisterPayload{"Alice", "Alice <alice@example.com>", "password123"}, "email must be a valid email address"},
		{"email whitespace", RegisterPayload{"Alice", " alice@example.com ", "password123"}, ""},
		{"empty password", RegisterPayload{"Alice", "alice@example.com", ""}, "password must contain at least 8 characters"},
		{"blank password", RegisterPayload{"Alice", "alice@example.com", "        "}, "password must contain at least 8 characters"},
		{"short password", RegisterPayload{"Alice", "alice@example.com", "1234567"}, "password must contain at least 8 characters"},
		{"minimum password", RegisterPayload{"Alice", "alice@example.com", "12345678"}, ""},
		{"maximum password", RegisterPayload{"Alice", "alice@example.com", strings.Repeat("a", 72)}, ""},
		{"long password", RegisterPayload{"Alice", "alice@example.com", strings.Repeat("a", 73)}, "password must not exceed 72 bytes"},
		{"short unicode password", RegisterPayload{"Alice", "alice@example.com", strings.Repeat("é", 7)}, "password must contain at least 8 characters"},
		{"unicode byte limit", RegisterPayload{"Alice", "alice@example.com", strings.Repeat("é", 37)}, "password must not exceed 72 bytes"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parsed, errs := tt.payload.Validate()
			if tt.wantError == "" {
				if errs != nil {
					t.Fatalf("unexpected errors: %v", errs)
				}
				return
			}
			if parsed != (RegisterPayload{}) {
				t.Fatal("expected empty payload on validation failure")
			}
			if len(errs) != 1 || errs[0].Error() != tt.wantError {
				t.Fatalf("Validate() = %v, want one error %q", errs, tt.wantError)
			}
		})
	}
}
