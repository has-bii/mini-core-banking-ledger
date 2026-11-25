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

// NewService returns a session service configured with the given store and lifetime.
// The lifetime must be positive.
func NewService(store sessionStore, lifetime time.Duration) *Service {
	return &Service{store: store, lifetime: lifetime}
}

// Create issues a session for userID and returns it with its raw token.
// Persistence receives only the token's hash.
//
// On success, the unpadded URL-safe token has this shape:
//
//	"AAECAwQFBgcICQoLDA0ODxAREhMUFRYXGBkaGxwdHh8"
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

// GetByToken returns the active session for the raw token.
// It returns ErrNotFound if the token is malformed or the session is missing or expired.
func (s *Service) GetByToken(ctx context.Context, token string) (Session, error) {
	if !validToken(token) {
		return Session{}, ErrNotFound
	}
	return s.store.GetByTokenHash(ctx, hashToken(token))
}

// Revoke deletes the session for the raw token and succeeds if a well-formed token has no stored session.
// It returns ErrNotFound for a malformed token.
func (s *Service) Revoke(ctx context.Context, token string) error {
	if !validToken(token) {
		return ErrNotFound
	}
	return s.store.DeleteByTokenHash(ctx, hashToken(token))
}

// validToken checks for the canonical unpadded URL-safe base64 encoding of a 32-byte token.
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
