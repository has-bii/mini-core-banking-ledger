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
	err := r.Scan(&a.ID, &a.Name, &a.Type, &a.SortOrder, &a.Balance, &a.CreatedAt)
	return a, err
}

func (s *Store) FindAllByUserID(ctx context.Context, userID string, order validation.QueryOrder) ([]Account, error) {
	if err := order.Validate(); err != nil {
		return nil, fmt.Errorf("find all accounts: %w", err)
	}
	const q = `SELECT id, name, type, sort_order, balance, created_at FROM accounts
	WHERE user_id = $1
	ORDER BY sort_order `

	// SQL keywords cannot be parameters; Validate limits order to asc or desc.
	rows, err := s.db.Query(ctx, q+string(order), userID)
	if err != nil {
		return nil, fmt.Errorf("find all accounts: %w", err)
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (Account, error) {
		return scanAccount(r)
	})
}
