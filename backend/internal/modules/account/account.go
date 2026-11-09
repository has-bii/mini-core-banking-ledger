package account

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("account not found")

type Account struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Password  *string   `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
