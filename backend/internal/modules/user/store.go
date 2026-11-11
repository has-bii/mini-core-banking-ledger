package user

import (
	"backend/internal/platform/storage/postgres"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

const uniqueViolation = "23505"

type Store struct{ db postgres.Beginner }

func NewStore(db postgres.Beginner) *Store { return &Store{db: db} }

type scanner interface{ Scan(dest ...any) error }

func scanUser(r scanner) (User, error) {
	var u User
	err := r.Scan(&u.ID, &u.Name, &u.Email, &u.Password, &u.Image, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (s *Store) FindByID(ctx context.Context, id string) (User, error) {
	const q = `SELECT id, name, email, password, image, created_at, updated_at FROM users WHERE id = $1`

	u, err := scanUser(s.db.QueryRow(ctx, q, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return User{}, ErrNotFound
	}
	if err != nil {
		return User{}, fmt.Errorf("user by id: %w", err)
	}
	return u, nil
}

func (s *Store) FindByEmail(ctx context.Context, email string) (User, error) {
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
