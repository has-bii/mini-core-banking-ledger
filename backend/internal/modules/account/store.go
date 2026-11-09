package account

import (
	"backend/internal/platform/database"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Store struct{ db database.DBTX }

func NewStore(db database.DBTX) *Store { return &Store{db: db} }

type scanner interface{ Scan(dest ...any) error }

func scanAccount(r scanner) (Account, error) {
	var a Account
	err := r.Scan(&a.ID, &a.UserID, &a.Password, &a.CreatedAt, &a.UpdatedAt)
	return a, err
}

func (s *Store) Create(ctx context.Context, userID, hashedPassword string) (Account, error) {
	const q = `INSERT INTO accounts ("userId", password, "createdAt", "updatedAt")
		VALUES ($1, $2, now(), now())
		RETURNING id, "userId", password, "createdAt", "updatedAt"`

	a, err := scanAccount(s.db.QueryRow(ctx, q, userID, hashedPassword))
	if err != nil {
		return Account{}, fmt.Errorf("create account: %w", err)
	}
	return a, nil
}

func (s *Store) FindByUserID(ctx context.Context, userID string) (Account, error) {
	const q = `SELECT id, "userId", password, "createdAt", "updatedAt" FROM accounts WHERE "userId" = $1`

	a, err := scanAccount(s.db.QueryRow(ctx, q, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Account{}, ErrNotFound
	}
	if err != nil {
		return Account{}, fmt.Errorf("account by user id: %w", err)
	}
	return a, nil
}
