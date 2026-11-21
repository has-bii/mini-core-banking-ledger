package httpx

type HTTPError struct {
	Status  int
	Message string
	cause   error
}

func (he *HTTPError) Error() string { return he.Message }

func (he *HTTPError) Unwrap() error { return he.cause }

func NewHTTPError(status int, message string) error {
	return &HTTPError{Status: status, Message: message}
}
