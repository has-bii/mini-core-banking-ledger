package validation

import "strings"

const (
	CodeRequired      = "REQUIRED"
	CodeInvalidFormat = "INVALID_FORMAT"
	CodeMinLength     = "MIN_LENGTH"
	CodeMaxLength     = "MAX_LENGTH"
	CodeInvalidChoice = "INVALID_CHOICE"
	CodeInvalidType   = "INVALID_TYPE"
	CodeUnknownField  = "UNKNOWN_FIELD"
)

type FieldError struct {
	Field   string `json:"field"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Errors preserves field-level failures when validation crosses a Go error interface.
type Errors []FieldError

func (errs Errors) Error() string {
	messages := make([]string, len(errs))
	for i, err := range errs {
		messages[i] = err.Message
	}
	return strings.Join(messages, "\n")
}
