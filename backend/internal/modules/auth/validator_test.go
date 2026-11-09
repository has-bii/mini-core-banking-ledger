package auth_test

import (
	"slices"
	"strings"
	"testing"

	"backend/internal/modules/auth"
)

func TestValidation(t *testing.T) {
	for _, tt := range []struct {
		name     string
		email    string
		password string
		want     []error
	}{
		{"valid", "alice@example.com", "password", nil},
		{"trim email", " alice@example.com ", "password", nil},
		{"invalid email", "Alice <alice@example.com>", "password", []error{auth.ErrInvalidEmail}},
		{"empty password", "alice@example.com", "", []error{auth.ErrPasswordTooShort}},
		{"blank password", "alice@example.com", "        ", []error{auth.ErrPasswordTooShort}},
		{"seven characters", "alice@example.com", "1234567", []error{auth.ErrPasswordTooShort}},
		{"eight unicode characters", "alice@example.com", strings.Repeat("é", 8), nil},
		{"seven unicode characters", "alice@example.com", strings.Repeat("é", 7), []error{auth.ErrPasswordTooShort}},
		{"72 bytes", "alice@example.com", strings.Repeat("a", 72), nil},
		{"73 bytes", "alice@example.com", strings.Repeat("a", 73), []error{auth.ErrPasswordTooLong}},
		{"unicode byte limit", "alice@example.com", strings.Repeat("é", 37), []error{auth.ErrPasswordTooLong}},
		{"multiple errors", "", "", []error{auth.ErrInvalidEmail, auth.ErrPasswordTooShort}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			register := auth.RegisterDTO{Name: " Alice ", Email: tt.email, Password: tt.password}
			if got := register.Validate(); !slices.Equal(got, tt.want) {
				t.Errorf("Register.Validate() = %v, want %v", got, tt.want)
			}
			login := auth.LoginDTO{Email: tt.email, Password: tt.password}
			if got := login.Validate(); !slices.Equal(got, tt.want) {
				t.Errorf("Login.Validate() = %v, want %v", got, tt.want)
			}
			if register.Name != "Alice" || register.Email != strings.TrimSpace(tt.email) || login.Email != strings.TrimSpace(tt.email) {
				t.Error("name and email should be trimmed")
			}
			if register.Password != tt.password || login.Password != tt.password {
				t.Error("password must not be modified")
			}
		})
	}
}

func TestRegisterRequiresName(t *testing.T) {
	payload := auth.RegisterDTO{Name: " \t ", Email: "alice@example.com", Password: "password"}
	if got := payload.Validate(); !slices.Equal(got, []error{auth.ErrNameRequired}) {
		t.Errorf("Validate() = %v, want name required", got)
	}
}
