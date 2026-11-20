package session

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"
)

type sessionStore interface {
	Create(ctx context.Context, userID, tokenHash string, expiresAt time.Time) (Session, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (Session, error)
	DeleteByTokenHash(ctx context.Context, tokenHash string) error
}

type Service struct {
	store    sessionStore
	lifetime time.Duration
}

// NewService configures the lifetime for new sessions. The lifetime must be positive.
func NewService(store sessionStore, lifetime time.Duration) *Service {
	return &Service{store: store, lifetime: lifetime}
}

// Create returns the raw token only to the caller; persistence receives its hash.
func (s *Service) Create(ctx context.Context, userID string) (Session, string, error) {
	var random [32]byte
	if _, err := rand.Read(random[:]); err != nil {
		return Session{}, "", fmt.Errorf("generate session token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(random[:])
	session, err := s.store.Create(ctx, userID, hashToken(token), time.Now().UTC().Add(s.lifetime))
	if err != nil {
		return Session{}, "", err
	}
	return session, token, nil
}

func (s *Service) GetByToken(ctx context.Context, token string) (Session, error) {
	if !validToken(token) {
		return Session{}, ErrNotFound
	}
	return s.store.GetByTokenHash(ctx, hashToken(token))
}

func (s *Service) Revoke(ctx context.Context, token string) error {
	if !validToken(token) {
		return ErrNotFound
	}
	return s.store.DeleteByTokenHash(ctx, hashToken(token))
}

func validToken(token string) bool {
	if len(token) != 43 {
		return false
	}
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(token)
	return err == nil && len(decoded) == 32
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
