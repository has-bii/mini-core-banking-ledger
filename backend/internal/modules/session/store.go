package session

import (
	"context"
	"encoding/json/v2"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const redisSessionPrefix = "session:"

type Store struct{ redis *redis.Client }

// NewStore returns a session store backed by the given Redis client.
func NewStore(rds *redis.Client) *Store { return &Store{redis: rds} }

// Create stores and returns a session using a service-generated token hash,
// never the raw token. It wraps ErrInvalidUserID for an invalid UUID and
// ErrInvalidExpiration if expiresAt is not in the future. A duplicate hash
// wraps ErrTokenHashTaken without replacing the existing session.
func (s *Store) Create(ctx context.Context, userID, tokenHash string, expiresAt time.Time) (Session, error) {
	if err := uuid.Validate(userID); err != nil {
		return Session{}, fmt.Errorf("create session: %w", ErrInvalidUserID)
	}
	if !expiresAt.After(time.Now()) {
		return Session{}, fmt.Errorf("create session: %w", ErrInvalidExpiration)
	}

	id, err := uuid.NewV7()
	if err != nil {
		return Session{}, fmt.Errorf("generate session id: %w", err)
	}

	session := Session{
		ID:        id.String(),
		UserID:    userID,
		ExpiresAt: expiresAt.UTC(),
	}

	data, err := json.Marshal(session)
	if err != nil {
		return Session{}, fmt.Errorf("create session: %w", err)
	}

	ttl := time.Until(session.ExpiresAt)
	if ttl <= 0 {
		return Session{}, fmt.Errorf("create session: %w", ErrInvalidExpiration)
	}

	// Store the value and TTL atomically; never overwrite an existing token hash.
	created, err := s.redis.SetNX(ctx, redisSessionPrefix+tokenHash, data, ttl).Result()
	if err != nil {
		return Session{}, fmt.Errorf("create session: %w", err)
	}
	if !created {
		return Session{}, fmt.Errorf("create session: %w", ErrTokenHashTaken)
	}

	return session, nil
}

// GetByTokenHash looks up an opaque, service-generated token hash without normalizing it.
// It returns ErrNotFound if the session is missing or expired.
func (s *Store) GetByTokenHash(ctx context.Context, tokenHash string) (Session, error) {
	data, err := s.redis.Get(ctx, redisSessionPrefix+tokenHash).Bytes()
	if errors.Is(err, redis.Nil) {
		return Session{}, ErrNotFound
	}
	if err != nil {
		return Session{}, fmt.Errorf("session by token: %w", err)
	}

	var session Session
	if err := json.Unmarshal(data, &session); err != nil {
		return Session{}, fmt.Errorf("decode session: %w", err)
	}
	if session.ID == "" || session.UserID == "" || session.ExpiresAt.IsZero() {
		return Session{}, fmt.Errorf("decode session: missing required fields")
	}
	if !session.ExpiresAt.After(time.Now()) {
		return Session{}, ErrNotFound
	}
	return session, nil
}

// DeleteByTokenHash revokes the session for a service-generated token hash.
// It succeeds if the key is already absent.
func (s *Store) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	if err := s.redis.Del(ctx, redisSessionPrefix+tokenHash).Err(); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
