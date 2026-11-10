package validation

import (
	"github.com/google/uuid"
)

func ValidUUID(value string) bool {
	err := uuid.Validate(value)
	return err == nil
}
