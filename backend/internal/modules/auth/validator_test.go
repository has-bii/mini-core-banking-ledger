package auth_test

import (
	"backend/internal/modules/auth"
	"slices"
	"strings"
	"testing"
)

func TestRegisterDTO_Validate(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name     string
		payload  auth.RegisterDTO
		want     []error
		wantName string
		wantMail string
	}{
		{"valid and trimmed", auth.RegisterDTO{Name: " Alice ", Email: " alice@example.com ", Password: "password123"}, nil, "Alice", "alice@example.com"},
		{"missing fields", auth.RegisterDTO{}, []error{auth.ErrNameRequired, auth.ErrInvalidEmail, auth.ErrPasswordTooShort}, "", ""},
		{"blank name", auth.RegisterDTO{Name: " \t", Email: "alice@example.com", Password: "password123"}, []error{auth.ErrNameRequired}, "", "alice@example.com"},
		{"invalid email", auth.RegisterDTO{Name: "Alice", Email: "not-an-email", Password: "password123"}, []error{auth.ErrInvalidEmail}, "Alice", "not-an-email"},
		{"short password", auth.RegisterDTO{Name: "Alice", Email: "alice@example.com", Password: "1234567"}, []error{auth.ErrPasswordTooShort}, "Alice", "alice@example.com"},
		{"blank password", auth.RegisterDTO{Name: "Alice", Email: "alice@example.com", Password: "        "}, []error{auth.ErrPasswordTooShort}, "Alice", "alice@example.com"},
		{"eight characters", auth.RegisterDTO{Name: "Alice", Email: "alice@example.com", Password: "éééééééé"}, nil, "Alice", "alice@example.com"},
		{"seven multibyte characters", auth.RegisterDTO{Name: "Alice", Email: "alice@example.com", Password: "ééééééé"}, []error{auth.ErrPasswordTooShort}, "Alice", "alice@example.com"},
		{"72 bytes", auth.RegisterDTO{Name: "Alice", Email: "alice@example.com", Password: strings.Repeat("é", 36)}, nil, "Alice", "alice@example.com"},
		{"over 72 bytes", auth.RegisterDTO{Name: "Alice", Email: "alice@example.com", Password: strings.Repeat("é", 37)}, []error{auth.ErrPasswordTooLong}, "Alice", "alice@example.com"},
		{"password whitespace preserved", auth.RegisterDTO{Name: "Alice", Email: "alice@example.com", Password: " password "}, nil, "Alice", "alice@example.com"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			password := tt.payload.Password
			if got := tt.payload.Validate(); !slices.Equal(got, tt.want) {
				t.Fatalf("Validate() = %v, want %v", got, tt.want)
			}
			if tt.payload.Name != tt.wantName || tt.payload.Email != tt.wantMail || tt.payload.Password != password {
				t.Errorf("normalization = %+v, want name %q, email %q and unchanged password", tt.payload, tt.wantName, tt.wantMail)
			}
		})
	}
}

func TestLoginDTO_Validate(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name     string
		email    string
		password string
		want     []error
	}{
		{"valid and trimmed", " alice@example.com ", " password ", nil},
		{"missing fields", "", "", []error{auth.ErrInvalidEmail, auth.ErrPasswordTooShort}},
		{"invalid email", "not-an-email", "password123", []error{auth.ErrInvalidEmail}},
		{"blank password", "alice@example.com", "        ", []error{auth.ErrPasswordTooShort}},
		{"seven characters", "alice@example.com", "ééééééé", []error{auth.ErrPasswordTooShort}},
		{"eight characters", "alice@example.com", "éééééééé", nil},
		{"72 bytes", "alice@example.com", strings.Repeat("é", 36), nil},
		{"over 72 bytes", "alice@example.com", strings.Repeat("é", 37), []error{auth.ErrPasswordTooLong}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			payload := auth.LoginDTO{Email: tt.email, Password: tt.password}
			if got := payload.Validate(); !slices.Equal(got, tt.want) {
				t.Fatalf("Validate() = %v, want %v", got, tt.want)
			}
			if payload.Email != strings.TrimSpace(tt.email) || payload.Password != tt.password {
				t.Errorf("normalization = %+v, want trimmed email and unchanged password", payload)
			}
		})
	}
}
