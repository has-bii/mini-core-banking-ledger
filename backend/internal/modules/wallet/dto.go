package wallet

import (
	"errors"
	"fmt"
	"strings"

	pValidator "backend/internal/platform/validator"

	"github.com/go-playground/validator/v10"
)

var validate = func() *validator.Validate {
	v := validator.New(validator.WithRequiredStructEnabled())
	v.RegisterTagNameFunc(pValidator.UseJSONName)
	return v
}()

type CreateWalletDTO struct {
	Name       string `json:"name" validate:"required,min=3,max=32,alphaspace"`
	WalletType string `json:"wallet_type" validate:"required,oneof=sub_wallet savings"`
}

// Validate trims Name and WalletType in place before checking the field rules.
// The fields remain trimmed even if validation fails.
func (p *CreateWalletDTO) Validate() error {
	p.Name = strings.TrimSpace(p.Name)
	p.WalletType = strings.TrimSpace(p.WalletType)

	err := validate.Struct(p)
	if err == nil {
		return nil
	}

	if _, ok := errors.AsType[*validator.InvalidValidationError](err); ok {
		return fmt.Errorf("validate wallet: %w", err)
	}

	mapped := pValidator.FormatErrors(err)

	return pValidator.NewErrorValidation(&mapped)
}
