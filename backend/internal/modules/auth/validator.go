package auth

import (
	"errors"
	"net/mail"
	"strings"
	"unicode/utf8"
)

type RegisterPayload struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (p RegisterPayload) Validate() (RegisterPayload, []error) {
	p.Name = strings.TrimSpace(p.Name)
	p.Email = strings.TrimSpace(p.Email)
	var errs []error

	if p.Name == "" {
		errs = append(errs, errors.New("name is required"))
	}

	// ParseAddress also accepts display names; registration needs a bare address.
	address, err := mail.ParseAddress(p.Email)
	if err != nil || address.Address != p.Email {
		errs = append(errs, errors.New("email must be a valid email address"))
	}

	if strings.TrimSpace(p.Password) == "" || utf8.RuneCountInString(p.Password) < 8 {
		errs = append(errs, errors.New("password must contain at least 8 characters"))
	}
	// bcrypt rejects passwords longer than 72 bytes, not 72 characters.
	if len(p.Password) > 72 {
		errs = append(errs, errors.New("password must not exceed 72 bytes"))
	}

	if len(errs) > 0 {
		return RegisterPayload{}, errs
	}
	return p, nil
}
