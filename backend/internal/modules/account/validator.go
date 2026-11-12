package account

import (
	"errors"
	"strings"

	"github.com/go-playground/validator/v10"
)

var accountValidator = validator.New(validator.WithRequiredStructEnabled())

type CreateAccountDTO struct {
	Name string      `json:"name" validate:"required,min=3,max=32,alphaspace"`
	Type AccountType `json:"account_type" validate:"oneof=sub_account savings"`
}

func (p *CreateAccountDTO) Validate() []error {
	p.Name = strings.TrimSpace(p.Name)
	p.Type = AccountType(strings.TrimSpace(string(p.Type)))

	err := accountValidator.Struct(p)
	if err == nil {
		return nil
	}
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		return []error{err}
	}

	var errs []error
	for _, field := range validationErrors {
		switch field.StructField() {
		case "Name":
			if field.Tag() == "required" {
				errs = append(errs, ErrNameRequired)
			} else {
				errs = append(errs, ErrInvalidName)
			}
		case "Type":
			errs = append(errs, ErrInvalidType)
		}
	}
	return errs
}
