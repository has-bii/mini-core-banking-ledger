package auth

import (
	"backend/internal/platform/validation"
	"strings"

	"github.com/go-playground/validator/v10"
)

var authValidator = validator.New(validator.WithRequiredStructEnabled())

type RegisterDTO struct {
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

func (p *RegisterDTO) Validate() validation.Errors {
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

func (p *LoginDTO) Validate() validation.Errors {
	p.Email = strings.TrimSpace(p.Email)
	payload := *p
	payload.Password = passwordForValidation(p.Password)
	return authValidationErrors(authValidator.Struct(payload), p.Password)
}

func passwordForValidation(password string) string {
	// Normalize only the validation copy; never trim the actual password.
	if authValidator.Var(strings.TrimSpace(password), "required") != nil {
		return ""
	}
	return password
}

func authValidationErrors(err error, password string) validation.Errors {
	var errs validation.Errors
	passwordInvalid := false
	if err != nil {
		// Struct receives only the concrete, non-nil DTOs above.
		for _, field := range err.(validator.ValidationErrors) {
			switch field.StructField() {
			case "Name":
				errs = append(errs, validation.FieldError{Field: "name", Code: validation.CodeRequired, Message: "Name is required."})
			case "Email":
				if field.Tag() == "required" {
					errs = append(errs, validation.FieldError{Field: "email", Code: validation.CodeRequired, Message: "Email is required."})
				} else {
					errs = append(errs, validation.FieldError{Field: "email", Code: validation.CodeInvalidFormat, Message: "Enter a valid email address."})
				}
			case "Password":
				passwordInvalid = true
				if field.Tag() == "required" {
					errs = append(errs, validation.FieldError{Field: "password", Code: validation.CodeRequired, Message: "Password is required."})
				} else {
					errs = append(errs, validation.FieldError{Field: "password", Code: validation.CodeMinLength, Message: "Password must contain at least 8 characters."})
				}
			}
		}
	}
	// A byte slice makes the built-in max rule enforce bcrypt's byte limit.
	if !passwordInvalid && authValidator.Var([]byte(password), "max=72") != nil {
		errs = append(errs, validation.FieldError{Field: "password", Code: validation.CodeMaxLength, Message: "Password must not exceed 72 bytes."})
	}
	return errs
}
