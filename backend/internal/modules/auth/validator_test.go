package auth_test

import (
	"backend/internal/modules/auth"
	"backend/internal/platform/validation"
	"slices"
	"strings"
	"testing"
)

func TestRegisterDTO_Validate(t *testing.T) {
	t.Parallel()
	nameRequired := validation.FieldError{Field: "name", Code: "REQUIRED", Message: "Name is required."}
	emailRequired := validation.FieldError{Field: "email", Code: "REQUIRED", Message: "Email is required."}
	emailInvalid := validation.FieldError{Field: "email", Code: "INVALID_FORMAT", Message: "Enter a valid email address."}
	passwordRequired := validation.FieldError{Field: "password", Code: "REQUIRED", Message: "Password is required."}
	passwordShort := validation.FieldError{Field: "password", Code: "MIN_LENGTH", Message: "Password must contain at least 8 characters."}
	passwordLong := validation.FieldError{Field: "password", Code: "MAX_LENGTH", Message: "Password must not exceed 72 bytes."}
	for _, tt := range []struct {
		name     string
		payload  auth.RegisterDTO
		want     validation.Errors
		wantName string
		wantMail string
	}{
		{"valid and trimmed", auth.RegisterDTO{Name: " Alice ", Email: " alice@example.com ", Password: "password123"}, nil, "Alice", "alice@example.com"},
		{"missing fields", auth.RegisterDTO{}, validation.Errors{nameRequired, emailRequired, passwordRequired}, "", ""},
		{"blank name", auth.RegisterDTO{Name: " \t", Email: "alice@example.com", Password: "password123"}, validation.Errors{nameRequired}, "", "alice@example.com"},
		{"invalid email", auth.RegisterDTO{Name: "Alice", Email: "not-an-email", Password: "password123"}, validation.Errors{emailInvalid}, "Alice", "not-an-email"},
		{"short password", auth.RegisterDTO{Name: "Alice", Email: "alice@example.com", Password: "1234567"}, validation.Errors{passwordShort}, "Alice", "alice@example.com"},
		{"blank password", auth.RegisterDTO{Name: "Alice", Email: "alice@example.com", Password: "        "}, validation.Errors{passwordRequired}, "Alice", "alice@example.com"},
		{"blank password over byte limit", auth.RegisterDTO{Name: "Alice", Email: "alice@example.com", Password: strings.Repeat(" ", 73)}, validation.Errors{passwordRequired}, "Alice", "alice@example.com"},
		{"eight characters", auth.RegisterDTO{Name: "Alice", Email: "alice@example.com", Password: "éééééééé"}, nil, "Alice", "alice@example.com"},
		{"seven multibyte characters", auth.RegisterDTO{Name: "Alice", Email: "alice@example.com", Password: "ééééééé"}, validation.Errors{passwordShort}, "Alice", "alice@example.com"},
		{"72 bytes", auth.RegisterDTO{Name: "Alice", Email: "alice@example.com", Password: strings.Repeat("é", 36)}, nil, "Alice", "alice@example.com"},
		{"over 72 bytes", auth.RegisterDTO{Name: "Alice", Email: "alice@example.com", Password: strings.Repeat("é", 37)}, validation.Errors{passwordLong}, "Alice", "alice@example.com"},
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
	emailRequired := validation.FieldError{Field: "email", Code: "REQUIRED", Message: "Email is required."}
	emailInvalid := validation.FieldError{Field: "email", Code: "INVALID_FORMAT", Message: "Enter a valid email address."}
	passwordRequired := validation.FieldError{Field: "password", Code: "REQUIRED", Message: "Password is required."}
	passwordShort := validation.FieldError{Field: "password", Code: "MIN_LENGTH", Message: "Password must contain at least 8 characters."}
	passwordLong := validation.FieldError{Field: "password", Code: "MAX_LENGTH", Message: "Password must not exceed 72 bytes."}
	for _, tt := range []struct {
		name     string
		email    string
		password string
		want     validation.Errors
	}{
		{"valid and trimmed", " alice@example.com ", " password ", nil},
		{"missing fields", "", "", validation.Errors{emailRequired, passwordRequired}},
		{"invalid email", "not-an-email", "password123", validation.Errors{emailInvalid}},
		{"blank password", "alice@example.com", "        ", validation.Errors{passwordRequired}},
		{"seven characters", "alice@example.com", "ééééééé", validation.Errors{passwordShort}},
		{"eight characters", "alice@example.com", "éééééééé", nil},
		{"71 bytes", "alice@example.com", strings.Repeat("a", 71), nil},
		{"72 bytes", "alice@example.com", strings.Repeat("é", 36), nil},
		{"73 bytes", "alice@example.com", strings.Repeat("a", 73), validation.Errors{passwordLong}},
		{"over 72 bytes multibyte", "alice@example.com", strings.Repeat("é", 37), validation.Errors{passwordLong}},
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
