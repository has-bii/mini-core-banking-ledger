package auth

import (
	"slices"
	"strings"
	"testing"
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
		{"invalid email", "Alice <alice@example.com>", "password", []error{ErrInvalidEmail}},
		{"empty password", "alice@example.com", "", []error{ErrPasswordTooShort}},
		{"blank password", "alice@example.com", "        ", []error{ErrPasswordTooShort}},
		{"seven characters", "alice@example.com", "1234567", []error{ErrPasswordTooShort}},
		{"eight unicode characters", "alice@example.com", strings.Repeat("é", 8), nil},
		{"seven unicode characters", "alice@example.com", strings.Repeat("é", 7), []error{ErrPasswordTooShort}},
		{"72 bytes", "alice@example.com", strings.Repeat("a", 72), nil},
		{"73 bytes", "alice@example.com", strings.Repeat("a", 73), []error{ErrPasswordTooLong}},
		{"unicode byte limit", "alice@example.com", strings.Repeat("é", 37), []error{ErrPasswordTooLong}},
		{"multiple errors", "", "", []error{ErrInvalidEmail, ErrPasswordTooShort}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			register := RegisterDTO{Name: " Alice ", Email: tt.email, Password: tt.password}
			if got := register.Validate(); !slices.Equal(got, tt.want) {
				t.Errorf("Register.Validate() = %v, want %v", got, tt.want)
			}
			login := LoginDTO{Email: tt.email, Password: tt.password}
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
	payload := RegisterDTO{Name: " \t ", Email: "alice@example.com", Password: "password"}
	if got := payload.Validate(); !slices.Equal(got, []error{ErrNameRequired}) {
		t.Errorf("Validate() = %v, want name required", got)
	}
}
