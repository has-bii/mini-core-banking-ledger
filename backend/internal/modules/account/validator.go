package account

import (
	"regexp"
	"strings"
)

var accountNamePattern = regexp.MustCompile(`^[a-zA-Z ]{3,32}$`)

type CreateAccountDTO struct {
	Name string      `json:"name"`
	Type AccountType `json:"account_type"`
}

func (p *CreateAccountDTO) Validate() []error {
	p.Name = strings.TrimSpace(p.Name)
	p.Type = AccountType(strings.TrimSpace(string(p.Type)))

	var errs []error

	if p.Name == "" {
		errs = append(errs, ErrNameRequired)
	} else if !accountNamePattern.MatchString(p.Name) {
		errs = append(errs, ErrInvalidName)
	}

	switch p.Type {
	case SubAccount, Savings:
	default:
		errs = append(errs, ErrInvalidType)
	}

	return errs
}
