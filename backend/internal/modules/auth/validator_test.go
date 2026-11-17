package auth_test

import (
	"backend/internal/modules/auth"
	"backend/internal/platform/validator"
	"errors"
	"reflect"
	"strings"
	"testing"
)

func assertValidationErrors(t *testing.T, err error, want validator.ErrorField) {
	t.Helper()
	if want == nil {
		if err != nil {
			t.Fatalf("Validate() = %v, want nil", err)
		}
		return
	}
	var fields *validator.ErrorValidation
	if !errors.As(err, &fields) || fields.Errors == nil || !reflect.DeepEqual(*fields.Errors, want) {
		t.Fatalf("Validate() = %v, want %+v", err, want)
	}
}

func TestDTO_Validate(t *testing.T) {
	t.Parallel()
	// Keep whitespace, character-count, and bcrypt byte-limit regressions together.
	for _, tt := range []struct {
		name, password string
		want           validator.ErrorField
	}{
		{"valid", "password123", nil},
		{"missing password", "", validator.ErrorField{"password": {"This field is required"}}},
		{"blank password", "        ", validator.ErrorField{"password": {"This field is required"}}},
		{"blank password over byte limit", strings.Repeat(" ", 73), validator.ErrorField{"password": {"This field is required"}}},
		{"seven characters", "1234567", validator.ErrorField{"password": {"Must contain at least 8 characters"}}},
		{"seven multibyte characters", "ééééééé", validator.ErrorField{"password": {"Must contain at least 8 characters"}}},
		{"eight multibyte characters", "éééééééé", nil},
		{"71 bytes", strings.Repeat("a", 71), nil},
		{"72 bytes", strings.Repeat("é", 36), nil},
		{"73 bytes", strings.Repeat("a", 73), validator.ErrorField{"password": {"Password must not exceed 72 bytes."}}},
		{"over 72 bytes multibyte", strings.Repeat("é", 37), validator.ErrorField{"password": {"Password must not exceed 72 bytes."}}},
		{"byte limit includes whitespace", strings.Repeat("a", 72) + " ", validator.ErrorField{"password": {"Password must not exceed 72 bytes."}}},
		{"password whitespace preserved", " 123456 ", nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			register := auth.RegisterDTO{Name: " Alice ", Email: " alice@example.com ", Password: tt.password}
			login := auth.LoginDTO{Email: " alice@example.com ", Password: tt.password}
			assertValidationErrors(t, register.Validate(), tt.want)
			assertValidationErrors(t, login.Validate(), tt.want)
			if register.Name != "Alice" || register.Email != "alice@example.com" || login.Email != "alice@example.com" || register.Password != tt.password || login.Password != tt.password {
				t.Fatal("expected trimmed name/email and unchanged passwords")
			}
		})
	}
}

func TestDTO_RequiredFields(t *testing.T) {
	t.Parallel()
	register := auth.RegisterDTO{Name: " \t", Email: " "}
	login := auth.LoginDTO{Email: " "}
	assertValidationErrors(t, register.Validate(), validator.ErrorField{"name": {"This field is required"}, "email": {"This field is required"}, "password": {"This field is required"}})
	assertValidationErrors(t, login.Validate(), validator.ErrorField{"email": {"This field is required"}, "password": {"This field is required"}})
}

func TestDTO_EmailValidation(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, email string
		valid       bool
	}{
		{"plain address", "alice@example.com", true},
		{"plus address", "alice+banking@example.com", true},
		{"quoted local part", `"alice"@example.com`, true},
		{"unqualified domain", "alice@localhost", false},
		{"IP literal domain", "alice@[127.0.0.1]", false},
		{"display name", "Alice <alice@example.com>", false},
		{"invalid email", "not-an-email", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			register := auth.RegisterDTO{Name: "Alice", Email: tt.email, Password: "password123"}
			login := auth.LoginDTO{Email: tt.email, Password: "password123"}
			var want validator.ErrorField
			if !tt.valid {
				want = validator.ErrorField{"email": {"Must be a valid email address"}}
			}
			assertValidationErrors(t, register.Validate(), want)
			assertValidationErrors(t, login.Validate(), want)
		})
	}
}
