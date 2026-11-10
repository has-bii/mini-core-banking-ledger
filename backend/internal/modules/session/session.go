package session

import (
	"errors"
	"time"
)

var ErrNotFound = errors.New("session not found or expired")

const Lifetime = 24 * time.Hour

type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}
