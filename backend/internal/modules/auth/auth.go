package auth

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUnauthenticated    = errors.New("authentication required")
	ErrNameRequired       = errors.New("name is required")
	ErrInvalidEmail       = errors.New("email must be a valid email address")
	ErrPasswordTooShort   = errors.New("password must contain at least 8 characters")
	ErrPasswordTooLong    = errors.New("password is too long; please use a shorter password")
)
