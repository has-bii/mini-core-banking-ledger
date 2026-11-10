package account

import (
	"errors"
	"time"
)

var (
	ErrNotFound         = errors.New("account not found")
	ErrSortOrderTaken   = errors.New("account sort order already taken")
	ErrInvalidSortOrder = errors.New("account sort order must be non-negative")
	ErrInvalidType      = errors.New("invalid account type")
)

type AccountType string

const (
	Main       AccountType = "main"
	SubAccount AccountType = "sub_account"
	Savings    AccountType = "savings"
)

type Account struct {
	ID        string      `json:"id"`
	UserID    string      `json:"user_id"`
	Name      string      `json:"name"`
	Type      AccountType `json:"account_type"`
	SortOrder int16       `json:"sort_order"`
	Balance   int64       `json:"balance"`
	CreatedAt time.Time   `json:"created_at"`
}
