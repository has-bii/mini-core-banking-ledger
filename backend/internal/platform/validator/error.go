package validator

type ErrorField = map[string][]string

type ErrorValidation struct {
	Errors *ErrorField `json:"errors,omitzero"`
}

// Error returns a summary without field-level details.
func (ev *ErrorValidation) Error() string {
	return "One or more fields are invalid. Check the field errors for details."
}

// NewErrorValidation returns an error that carries the supplied field details.
func NewErrorValidation(errs *ErrorField) error {
	return &ErrorValidation{Errors: errs}
}
