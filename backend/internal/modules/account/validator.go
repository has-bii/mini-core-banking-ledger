package account

import (
	"backend/internal/platform/validation"
	"strings"

	"github.com/go-playground/validator/v10"
)

var accountValidator = validator.New(validator.WithRequiredStructEnabled())

type CreateAccountDTO struct {
	Name string      `json:"name" validate:"required,min=3,max=32,alphaspace"`
	Type AccountType `json:"account_type" validate:"required,oneof=sub_account savings"`
}

func (p *CreateAccountDTO) Validate() validation.Errors {
	p.Name = strings.TrimSpace(p.Name)
	p.Type = AccountType(strings.TrimSpace(string(p.Type)))
	err := accountValidator.Struct(p)
	if err == nil {
		return nil
	}
	var errs validation.Errors
	// Struct receives the concrete, non-nil DTO above.
	for _, field := range err.(validator.ValidationErrors) {
		switch field.StructField() {
		case "Name":
			switch field.Tag() {
			case "required":
				errs = append(errs, validation.FieldError{Field: "name", Code: validation.CodeRequired, Message: "Name is required."})
			case "min":
				errs = append(errs, validation.FieldError{Field: "name", Code: validation.CodeMinLength, Message: "Name must contain at least 3 characters."})
			case "max":
				errs = append(errs, validation.FieldError{Field: "name", Code: validation.CodeMaxLength, Message: "Name must contain at most 32 characters."})
			case "alphaspace":
				errs = append(errs, validation.FieldError{Field: "name", Code: validation.CodeInvalidFormat, Message: "Name must contain only ASCII letters or spaces."})
			}
		case "Type":
			if field.Tag() == "required" {
				errs = append(errs, validation.FieldError{Field: "account_type", Code: validation.CodeRequired, Message: "Account type is required."})
			} else {
				errs = append(errs, validation.FieldError{Field: "account_type", Code: validation.CodeInvalidChoice, Message: `Account type must be "sub_account" or "savings".`})
			}
		}
	}
	return errs
}
