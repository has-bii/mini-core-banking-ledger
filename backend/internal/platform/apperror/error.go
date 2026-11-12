package apperror

// Error describes a client-safe application failure. Details must be JSON-serializable.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
	cause   error
}

func New(code, message string, details any, cause error) *Error {
	return &Error{Code: code, Message: message, Details: details, cause: cause}
}

func (e *Error) Error() string { return e.Message }

func (e *Error) Unwrap() error { return e.cause }
