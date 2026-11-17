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

func (p *LoginDTO) Validate() error {
	p.Email = strings.TrimSpace(p.Email)
	payload := *p
	payload.Password = passwordForValidation(p.Password)
	return authValidationErrors(authValidator.Struct(payload), p.Password)
}

func passwordForValidation(password string) string {
	// Normalize only the validation copy; never trim the actual password.
	if strings.TrimSpace(password) == "" {
		return ""
	}
	return password
}

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
		fields["password"] = []string{"Password must not exceed 72 bytes."}
	}
	if len(fields) == 0 {
		return nil
	}
	return pValidator.NewErrorValidation(&fields)
}
