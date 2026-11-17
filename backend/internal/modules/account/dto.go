package account

import (
	"errors"
	"fmt"
	"strings"

	pValidator "backend/internal/platform/v2/validator"

	"github.com/go-playground/validator/v10"
)

var validate = func() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterTagNameFunc(pValidator.UseJSONName)
	return v
}()

type CreateAccountDTO struct {
	Name        string `json:"name" validate:"required,min=3,max=32,alphaspace"`
	AccountType string `json:"account_type" validate:"required,oneof=sub_account savings"`
}

func (p *CreateAccountDTO) Validate() error {
	p.Name = strings.TrimSpace(p.Name)
	p.AccountType = strings.TrimSpace(p.AccountType)

	err := validate.Struct(p)
	if err == nil {
		return nil
	}

	if _, ok := errors.AsType[*validator.InvalidValidationError](err); ok {
		return fmt.Errorf("validate account: %w", err)
	}

	mapped := pValidator.FormatErrors(err)

	return pValidator.NewErrorValidation(&mapped)
}
