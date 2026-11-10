package validation

import (
	"errors"
	"strings"
)

type QueryOrder string

const (
	ASC  QueryOrder = "asc"
	DESC QueryOrder = "desc"
)

var (
	ErrValidationError = errors.New("order must be either 'asc' or 'desc'")
)

func (q *QueryOrder) Validate() error {
	if q == nil {
		return ErrValidationError
	}

	trimmed := strings.TrimSpace(string(*q))
	normalized := QueryOrder(strings.ToLower(trimmed))

	switch normalized {
	case ASC, DESC:
		*q = normalized
		return nil
	default:
		return ErrValidationError
	}
}
