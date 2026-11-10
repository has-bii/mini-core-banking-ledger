package account

import (
	"backend/internal/platform/storage/postgres"
	"backend/internal/platform/validation"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Store struct{ db postgres.DBTX }

func NewStore(db postgres.DBTX) *Store { return &Store{db: db} }

type scanner interface{ Scan(dest ...any) error }

func scanAccount(r scanner) (Account, error) {
	var a Account
	err := r.Scan(&a.ID, &a.UserID, &a.Name, &a.Type, &a.SortOrder, &a.Balance, &a.CreatedAt)
	return a, err
}

func (s *Store) ListByUserID(ctx context.Context, userID string) ([]Account, error) {
	if !validation.IsValidUUID(userID) {
		return nil, validation.ErrInvalidUUID
	}

	const q = `SELECT id, user_id, name, type, sort_order, balance, created_at FROM accounts
	WHERE user_id = $1
	ORDER BY sort_order ASC`

	rows, err := s.db.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list by user id: %w", err)
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (Account, error) {
		return scanAccount(r)
	})
}
