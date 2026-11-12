package account

import (
	"backend/internal/platform/validation"
	"regexp"
	"strings"
	"unicode/utf8"
)

var accountNamePattern = regexp.MustCompile(`^[A-Za-z ]+$`)

type CreateAccountDTO struct {
	Name string      `json:"name"`
	Type AccountType `json:"account_type"`
}

func (p *CreateAccountDTO) Validate() validation.Errors {
	p.Name = strings.TrimSpace(p.Name)
	p.Type = AccountType(strings.TrimSpace(string(p.Type)))
	var errs validation.Errors
	switch {
	case p.Name == "":
		errs = append(errs, validation.FieldError{Field: "name", Code: validation.CodeRequired, Message: "Name is required."})
	case utf8.RuneCountInString(p.Name) < 3:
		errs = append(errs, validation.FieldError{Field: "name", Code: validation.CodeMinLength, Message: "Name must contain at least 3 characters."})
	case utf8.RuneCountInString(p.Name) > 32:
		errs = append(errs, validation.FieldError{Field: "name", Code: validation.CodeMaxLength, Message: "Name must contain at most 32 characters."})
	case !accountNamePattern.MatchString(p.Name):
		errs = append(errs, validation.FieldError{Field: "name", Code: validation.CodeInvalidFormat, Message: "Name must contain only ASCII letters or spaces."})
	}

	if p.Type == "" {
		errs = append(errs, validation.FieldError{Field: "account_type", Code: validation.CodeRequired, Message: "Account type is required."})
	} else if p.Type != SubAccount && p.Type != Savings {
		errs = append(errs, validation.FieldError{Field: "account_type", Code: validation.CodeInvalidChoice, Message: `Account type must be "sub_account" or "savings".`})
	}
	return errs
}
