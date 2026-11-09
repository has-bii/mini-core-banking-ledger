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
	Create(context.Context, string, string, time.Time) (Session, error)
	FindByTokenHash(context.Context, string) (Session, error)
	DeleteByTokenHash(context.Context, string) error
}

type Service struct{ store sessionStore }

func NewService(store sessionStore) *Service { return &Service{store: store} }

// Create returns the raw token only to the caller; persistence receives its hash.
func (s *Service) Create(ctx context.Context, userID string) (Session, string, error) {
	var random [32]byte
	if _, err := rand.Read(random[:]); err != nil {
		return Session{}, "", fmt.Errorf("generate session token: %w", err)
	}
	token := base64.RawURLEncoding.EncodeToString(random[:])
	session, err := s.store.Create(ctx, userID, hashToken(token), time.Now().UTC().Add(Lifetime))
	if err != nil {
		return Session{}, "", fmt.Errorf("create session: %w", err)
	}
	return session, token, nil
}

func (s *Service) FindByToken(ctx context.Context, token string) (Session, error) {
	if !validToken(token) {
		return Session{}, ErrNotFound
	}
	return s.store.FindByTokenHash(ctx, hashToken(token))
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
