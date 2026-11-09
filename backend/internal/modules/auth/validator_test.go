package auth_test

import (
	"strings"
	"testing"

	"backend/internal/modules/auth"
)

func TestRegisterDTOValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		userName string
		email    string
		password string
		errMsg   string
	}{
		{"valid", "Alice", "alice@example.com", "password123", ""},
		{"blank name", " \t ", "alice@example.com", "password123", "name is required"},
		{"empty email", "Alice", "", "password123", "email must be a valid email address"},
		{"invalid email", "Alice", "alice", "password123", "email must be a valid email address"},
		{"display name", "Alice", "Alice <alice@example.com>", "password123", "email must be a valid email address"},
		{"empty password", "Alice", "alice@example.com", "", "password must contain at least 8 characters"},
		{"blank password", "Alice", "alice@example.com", "        ", "password must contain at least 8 characters"},
		{"short password", "Alice", "alice@example.com", "1234567", "password must contain at least 8 characters"},
		{"minimum password", "Alice", "alice@example.com", "12345678", ""},
		{"short unicode password", "Alice", "alice@example.com", strings.Repeat("界", 7), "password must contain at least 8 characters"},
		{"unicode password", "Alice", "alice@example.com", strings.Repeat("界", 8), ""},
		{"maximum password bytes", "Alice", "alice@example.com", strings.Repeat("a", 72), ""},
		{"too many password bytes", "Alice", "alice@example.com", strings.Repeat("a", 73), "password is too long; please use a shorter password"},
		{"unicode maximum bytes", "Alice", "alice@example.com", strings.Repeat("界", 24), ""},
		{"unicode over byte limit", "Alice", "alice@example.com", strings.Repeat("界", 25), "password is too long; please use a shorter password"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := auth.RegisterDTO{Name: tt.userName, Email: tt.email, Password: tt.password}
			errs := payload.Validate()
			if tt.errMsg == "" {
				if errs != nil {
					t.Fatalf("Validate() = %v; want nil", errs)
				}
				return
			}
			if len(errs) != 1 || errs[0].Error() != tt.errMsg {
				t.Fatalf("errors = %v; want %q", errs, tt.errMsg)
			}
		})
	}
}

func TestRegisterDTONormalization(t *testing.T) {
	payload := auth.RegisterDTO{Name: " Alice \t", Email: " alice@example.com \n", Password: " password123 "}
	if errs := payload.Validate(); errs != nil {
		t.Fatal(errs)
	}
	if payload.Name != "Alice" || payload.Email != "alice@example.com" || payload.Password != " password123 " {
		t.Fatalf("unexpected normalization: %+v", payload)
	}
}

func TestRegisterDTONormalizationOnFailure(t *testing.T) {
	payload := auth.RegisterDTO{Name: " Alice \t", Email: " invalid \n", Password: "short"}
	if errs := payload.Validate(); len(errs) != 2 {
		t.Fatalf("errors = %v; want 2 errors", errs)
	}
	if payload.Name != "Alice" || payload.Email != "invalid" || payload.Password != "short" {
		t.Fatalf("unexpected normalization: %+v", payload)
	}
}

func TestLoginDTOValidate(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		email    string
		password string
		errMsg   string
	}{
		{"valid", "alice@example.com", "password123", ""},
		{"empty email", "", "password123", "email must be a valid email address"},
		{"blank email", " \t ", "password123", "email must be a valid email address"},
		{"invalid email", "alice", "password123", "email must be a valid email address"},
		{"display name", "Alice <alice@example.com>", "password123", "email must be a valid email address"},
		{"empty password", "alice@example.com", "", "password must contain at least 8 characters"},
		{"blank password", "alice@example.com", "        ", "password must contain at least 8 characters"},
		{"short password", "alice@example.com", "1234567", "password must contain at least 8 characters"},
		{"minimum password", "alice@example.com", "12345678", ""},
		{"short unicode password", "alice@example.com", strings.Repeat("界", 7), "password must contain at least 8 characters"},
		{"unicode password", "alice@example.com", strings.Repeat("界", 8), ""},
		{"maximum password bytes", "alice@example.com", strings.Repeat("a", 72), ""},
		{"too many password bytes", "alice@example.com", strings.Repeat("a", 73), "password is too long; please use a shorter password"},
		{"unicode maximum bytes", "alice@example.com", strings.Repeat("界", 24), ""},
		{"unicode over byte limit", "alice@example.com", strings.Repeat("界", 25), "password is too long; please use a shorter password"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			payload := auth.LoginDTO{Email: tt.email, Password: tt.password}
			errs := payload.Validate()
			if tt.errMsg == "" {
				if errs != nil {
					t.Fatalf("Validate() = %v; want nil", errs)
				}
				return
			}
			if len(errs) != 1 || errs[0].Error() != tt.errMsg {
				t.Fatalf("errors = %v; want %q", errs, tt.errMsg)
			}
		})
	}
}

func TestLoginDTONormalization(t *testing.T) {
	t.Parallel()
	for _, email := range []string{" alice@example.com \n", " invalid \n"} {
		t.Run(strings.TrimSpace(email), func(t *testing.T) {
			payload := auth.LoginDTO{Email: email, Password: " password123 "}
			payload.Validate()
			if payload.Email != strings.TrimSpace(email) || payload.Password != " password123 " {
				t.Fatalf("unexpected normalization: %+v", payload)
			}
		})
	}
}

func TestLoginDTOCollectsErrors(t *testing.T) {
	t.Parallel()
	payload := auth.LoginDTO{}
	errs := payload.Validate()
	if len(errs) != 2 {
		t.Fatalf("got %d errors, want 2: %v", len(errs), errs)
	}
}

func TestRegisterDTOCollectsErrors(t *testing.T) {
	payload := auth.RegisterDTO{}
	errs := payload.Validate()
	if len(errs) != 3 {
		t.Fatalf("got %d errors, want 3: %v", len(errs), errs)
	}
}
