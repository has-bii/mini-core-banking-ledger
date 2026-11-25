package httpx

type HTTPError struct {
	Status  int
	Message string
	cause   error
}

// Error returns the client-facing message without exposing the cause.
func (he *HTTPError) Error() string { return he.Message }

// Unwrap returns the underlying cause, or nil if none was supplied.
func (he *HTTPError) Unwrap() error { return he.cause }

// NewHTTPError returns an HTTPError with the given status and public message
// and no underlying cause.
func NewHTTPError(status int, message string) error {
	return WrapHTTPError(status, message, nil)
}

// WrapHTTPError returns an HTTPError with the given status and public message.
// The cause is available through unwrapping but is not part of the message.
func WrapHTTPError(status int, message string, cause error) error {
	return &HTTPError{Status: status, Message: message, cause: cause}
}
