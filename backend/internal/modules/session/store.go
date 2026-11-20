package session

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const redisSessionPrefix = "session:"

type Store struct{ redis *redis.Client }

func NewStore(rds *redis.Client) *Store { return &Store{redis: rds} }

// Create requires a service-generated token hash; it never persists raw tokens.
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

// DeleteByTokenHash revokes a service-generated token hash; missing keys are a success.
func (s *Store) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	if err := s.redis.Del(ctx, redisSessionPrefix+tokenHash).Err(); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
