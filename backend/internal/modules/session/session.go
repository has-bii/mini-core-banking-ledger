package session

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("session not found or expired")

const Lifetime = 24 * time.Hour

type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"userId"`
	ExpiresAt time.Time `json:"expiresAt"`
}
