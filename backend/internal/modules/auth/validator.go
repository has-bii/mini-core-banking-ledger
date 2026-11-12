package auth

import (
	"strings"
	"unicode/utf8"

	"backend/internal/platform/validation"
)

type RegisterDTO struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (p *RegisterDTO) Validate() validation.Errors {
	p.Name = strings.TrimSpace(p.Name)
	p.Email = strings.TrimSpace(p.Email)
	var errs validation.Errors
	if p.Name == "" {
		errs = append(errs, validation.FieldError{Field: "name", Code: validation.CodeRequired, Message: "Name is required."})
	}
	return append(errs, validateCredentials(p.Email, p.Password)...)
}

type LoginDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (p *LoginDTO) Validate() validation.Errors {
	p.Email = strings.TrimSpace(p.Email)
	return validateCredentials(p.Email, p.Password)
}

func validateCredentials(email, password string) validation.Errors {
	var errs validation.Errors
	if email == "" {
		errs = append(errs, validation.FieldError{Field: "email", Code: validation.CodeRequired, Message: "Email is required."})
	} else if !validation.IsValidEmail(email) {
		errs = append(errs, validation.FieldError{Field: "email", Code: validation.CodeInvalidFormat, Message: "Enter a valid email address."})
	}

	switch {
	case strings.TrimSpace(password) == "":
		errs = append(errs, validation.FieldError{Field: "password", Code: validation.CodeRequired, Message: "Password is required."})
	case utf8.RuneCountInString(password) < 8:
		errs = append(errs, validation.FieldError{Field: "password", Code: validation.CodeMinLength, Message: "Password must contain at least 8 characters."})
	// bcrypt rejects passwords longer than 72 bytes, not 72 characters.
	case len(password) > 72:
		errs = append(errs, validation.FieldError{Field: "password", Code: validation.CodeMaxLength, Message: "Password must not exceed 72 bytes."})
	}
	return errs
}
