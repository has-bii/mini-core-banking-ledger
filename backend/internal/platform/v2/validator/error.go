package validator

type FieldError struct {
	Field   string   `json:"field"`
	Message []string `json:"message"`
}

type ErrorValidation struct {
	Errors []FieldError `json:"errors,omitempty"`
}
