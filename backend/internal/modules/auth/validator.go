package auth

import (
	"errors"
	"strings"
	"unicode/utf8"

	"backend/internal/platform/identifier"
)

type RegisterDTO struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (p *RegisterDTO) Validate() []error {
	p.Name = strings.TrimSpace(p.Name)
	p.Email = strings.TrimSpace(p.Email)
	var errs []error

	if p.Name == "" {
		errs = append(errs, errors.New("name is required"))
	}

	if !identifier.IsValidEmail(p.Email) {
		errs = append(errs, errors.New("email must be a valid email address"))
	}

	if strings.TrimSpace(p.Password) == "" || utf8.RuneCountInString(p.Password) < 8 {
		errs = append(errs, errors.New("password must contain at least 8 characters"))
	}
	// bcrypt rejects passwords longer than 72 bytes, not 72 characters.
	if len(p.Password) > 72 {
		errs = append(errs, errors.New("password is too long; please use a shorter password"))
	}

	return errs
}

type LoginDTO struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (p *LoginDTO) Validate() []error {
	p.Email = strings.TrimSpace(p.Email)

	var errs []error

	if !identifier.IsValidEmail(p.Email) {
		errs = append(errs, errors.New("email must be a valid email address"))
	}

	if strings.TrimSpace(p.Password) == "" || utf8.RuneCountInString(p.Password) < 8 {
		errs = append(errs, errors.New("password must contain at least 8 characters"))
	}

	// bcrypt rejects passwords longer than 72 bytes, not 72 characters.
	if len(p.Password) > 72 {
		errs = append(errs, errors.New("password is too long; please use a shorter password"))
	}

	return errs
}
