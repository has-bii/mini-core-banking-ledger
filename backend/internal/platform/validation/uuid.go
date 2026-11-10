package validation

import (
	"errors"

	"github.com/google/uuid"
)

var ErrInvalidUUID = errors.New("Invalid UUID")

func IsValidUUID(value string) bool {
	err := uuid.Validate(value)
	return err == nil
}
