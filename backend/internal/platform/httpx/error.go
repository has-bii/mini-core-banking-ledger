package httpx

type HTTPError struct {
	Status  int
	Message string
	cause   error
}

func (he *HTTPError) Error() string { return he.Message }

func (he *HTTPError) Unwrap() error { return he.cause }

func NewHTTPError(status int, message string) error {
	return WrapHTTPError(status, message, nil)
}

// WrapHTTPError keeps the cause separate from the client-facing message.
func WrapHTTPError(status int, message string, cause error) error {
	return &HTTPError{Status: status, Message: message, cause: cause}
}
