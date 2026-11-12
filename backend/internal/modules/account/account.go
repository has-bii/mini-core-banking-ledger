package account

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("account not found")

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
