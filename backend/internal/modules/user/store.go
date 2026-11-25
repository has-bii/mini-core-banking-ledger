package user

import (
	"backend/internal/platform/postgres"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const uniqueViolation = "23505"

type Store struct{ db postgres.DBTX }

// NewStore returns a user store backed by db.
func NewStore(db postgres.DBTX) *Store { return &Store{db: db} }

type scanner interface{ Scan(dest ...any) error }

func scanUser(r scanner) (User, error) {
	var u User
	err := r.Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.Image, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

// GetByID returns the user with the given ID. It returns ErrNotFound if no user
// matches and wraps ErrInvalidUserID if userID is not a valid UUID.
func (s *Store) GetByID(ctx context.Context, userID string) (User, error) {
	if err := uuid.Validate(userID); err != nil {
		return User{}, fmt.Errorf("user by id: %w", ErrInvalidUserID)
	}

	const q = `SELECT id, name, email, password, image, created_at, updated_at FROM users WHERE id = $1`

	u, err := scanUser(s.db.QueryRow(ctx, q, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("user by id: %w", err)
	}
	return u, nil
}

// GetByEmail returns the user matching email without normalizing it.
// It returns ErrNotFound if no user matches.
func (s *Store) GetByEmail(ctx context.Context, email string) (User, error) {
	const q = `SELECT id, name, email, password, image, created_at, updated_at FROM users WHERE email = $1`

	u, err := scanUser(s.db.QueryRow(ctx, q, email))
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("user by email: %w", err)
	}
	return u, nil
}

// Create inserts and returns a user and provisions their main wallet.
// It requires normalized, validated name and email values and an already-hashed
// password. It returns ErrEmailTaken if the email is already in use.
func (s *Store) Create(ctx context.Context, name, email, password string) (User, error) {
	const q = `INSERT INTO users (name, email, password)
		VALUES ($1, $2, $3)
		RETURNING id, name, email, password, image, created_at, updated_at`

	u, err := scanUser(s.db.QueryRow(ctx, q, name, email, password))
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolation && pgErr.ConstraintName == "users_email_key" {
			return User{}, ErrEmailTaken
		}
		return User{}, fmt.Errorf("create user: %w", err)
	}
	return u, nil
}
