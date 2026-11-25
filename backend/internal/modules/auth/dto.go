package auth

import (
	pValidator "backend/internal/platform/validator"
	"errors"
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

var authValidator = func() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterTagNameFunc(pValidator.UseJSONName)
	return v
}()

type RegisterDTO struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// Validate trims Name and Email in place, even if validation fails.
// It leaves Password unchanged and returns a [pValidator.ErrorValidation]
// with messages keyed by JSON field name for invalid input.
func (p *RegisterDTO) Validate() error {
	p.Name = strings.TrimSpace(p.Name)
	p.Email = strings.TrimSpace(p.Email)
	payload := *p
	payload.Password = passwordForValidation(p.Password)
	return authValidationErrors(authValidator.Struct(payload), p.Password)
}

type LoginDTO struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// Validate trims Email in place, even if validation fails.
// It leaves Password unchanged and returns a [pValidator.ErrorValidation]
// with messages keyed by JSON field name for invalid input.
func (p *LoginDTO) Validate() error {
	p.Email = strings.TrimSpace(p.Email)
	payload := *p
	payload.Password = passwordForValidation(p.Password)
	return authValidationErrors(authValidator.Struct(payload), p.Password)
}

// passwordForValidation treats whitespace-only passwords as empty for the
// required-field check and preserves all other passwords as entered.
func passwordForValidation(password string) string {
	// Normalize only the validation copy; never trim the actual password.
	if strings.TrimSpace(password) == "" {
		return ""
	}
	return password
}

// authValidationErrors returns JSON-named field errors for invalid input,
// wraps validator misuse errors, and returns nil if there are no violations.
func authValidationErrors(err error, password string) error {
	if _, ok := errors.AsType[*validator.InvalidValidationError](err); ok {
		return fmt.Errorf("validate auth: %w", err)
	}
	fields := pValidator.FormatErrors(err)
	// bcrypt limits bytes, while the min rule counts characters.
	if len(fields["password"]) == 0 && len(password) > 72 {
		if fields == nil {
			fields = make(pValidator.ErrorField)
		}
		fields["password"] = []string{"Password is too long"}
	}
	if len(fields) == 0 {
		return nil
	}
	return pValidator.NewErrorValidation(&fields)
}
