package validator

type ErrorField struct {
	Field   string   `json:"field"`
	Message []string `json:"message"`
}

func NewErrorField(field string, message []string) *ErrorField {
	return &ErrorField{Field: field, Message: message}
}

type ErrorValidation struct {
	Errors *[]ErrorField `json:"errors,omitempty"`
}

func (ev *ErrorValidation) Error() string { return "Validation Failed" } // TODO: change the returned string

func NewErrorValidation(errs *[]ErrorField) error {
	return &ErrorValidation{Errors: errs}
}
