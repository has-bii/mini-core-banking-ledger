package session

import (
	"backend/internal/platform/storage/postgres"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
)

type Store struct{ db postgres.DBTX }

func NewStore(db postgres.DBTX) *Store { return &Store{db: db} }

func (s *Store) Create(ctx context.Context, userID, tokenHash string, expiresAt time.Time) (Session, error) {
	const q = `INSERT INTO sessions ("userId", token, "expiresAt", "createdAt", "updatedAt")
		VALUES ($1, $2, $3, now(), now())
		RETURNING id, "userId", "expiresAt"`
	var session Session
	err := s.db.QueryRow(ctx, q, userID, tokenHash, expiresAt).Scan(&session.ID, &session.UserID, &session.ExpiresAt)
	if err != nil {
		return Session{}, fmt.Errorf("create session: %w", err)
	}
	return session, nil
}

func (s *Store) FindByTokenHash(ctx context.Context, tokenHash string) (Session, error) {
	const q = `SELECT id, "userId", "expiresAt" FROM sessions WHERE token = $1 AND "expiresAt" > now()`
	var session Session
	err := s.db.QueryRow(ctx, q, tokenHash).Scan(&session.ID, &session.UserID, &session.ExpiresAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return Session{}, ErrNotFound
	}
	if err != nil {
		return Session{}, fmt.Errorf("session by token: %w", err)
	}
	return session, nil
}

func (s *Store) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	_, err := s.db.Exec(ctx, `DELETE FROM sessions WHERE token = $1`, tokenHash)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
