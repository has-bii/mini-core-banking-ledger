package session

import (
	"errors"
	"time"
)

var (
	ErrNotFound          = errors.New("session not found or expired")
	ErrInvalidUserID     = errors.New("invalid user ID")
	ErrInvalidExpiration = errors.New("session expiration must be in the future")
	ErrTokenHashTaken    = errors.New("session token hash already exists")
)

const Lifetime = 24 * time.Hour

type Session struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	ExpiresAt time.Time `json:"expires_at"`
}
