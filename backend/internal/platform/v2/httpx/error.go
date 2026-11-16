package httpx

type HTTPError struct {
	Status  int
	Message string
}

func (he *HTTPError) Error() string { return he.Message }

func NewHTTPError(status int, message string) error {
	return &HTTPError{Status: status, Message: message}
}
