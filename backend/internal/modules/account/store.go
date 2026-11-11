package account

import (
	"backend/internal/platform/storage/postgres"
	"backend/internal/platform/validation"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
)

type Store struct{ db postgres.Beginner }

func NewStore(db postgres.Beginner) *Store { return &Store{db: db} }

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

func (s *Store) GetByID(ctx context.Context, ID, userID string) (Account, error) {
	if !validation.IsValidUUID(ID) {
		return Account{}, fmt.Errorf("account by id: ID %w", validation.ErrInvalidUUID)
	}
	if !validation.IsValidUUID(userID) {
		return Account{}, fmt.Errorf("account by id: userID %w", validation.ErrInvalidUUID)
	}

	const q = `SELECT id, user_id, name, type, sort_order, balance, created_at FROM accounts
	WHERE id = $1 AND user_id = $2`

	a, err := scanAccount(s.db.QueryRow(ctx, q, ID, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Account{}, ErrNotFound
	}
	if err != nil {
		return Account{}, fmt.Errorf("account by id: %w", err)
	}
	return a, nil
}

func (s *Store) Create(ctx context.Context, userID string, data *CreateAccountDTO) (Account, error) {
	if !validation.IsValidUUID(userID) {
		return Account{}, fmt.Errorf("create account: userID %w", validation.ErrInvalidUUID)
	}

	if errs := data.Validate(); len(errs) > 0 {
		return Account{}, ErrInvalidType
	}

	var a Account
	err := postgres.InTx(ctx, s.db, func(tx pgx.Tx) error {
		// Serialize sort-order allocation for this user until the transaction commits.
		var lockedUserID string
		err := tx.QueryRow(ctx, `SELECT id FROM users WHERE id = $1 FOR UPDATE`, userID).Scan(&lockedUserID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("lock account owner: %w", err)
		}

		var lastSortOrder int
		err = tx.QueryRow(ctx,
			`SELECT sort_order FROM accounts WHERE user_id = $1 ORDER BY sort_order DESC LIMIT 1`,
			userID).Scan(&lastSortOrder)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		err = tx.QueryRow(ctx,
			`INSERT INTO accounts (user_id, name, type, sort_order)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, name, type, sort_order, balance, created_at`,
			userID, data.Name, data.Type, lastSortOrder+1,
		).Scan(&a.ID, &a.UserID, &a.Name, &a.Type, &a.SortOrder, &a.Balance, &a.CreatedAt)

		if err != nil {
			return fmt.Errorf("create account: %w", err)
		}
		return nil
	})
	if err != nil {
		return Account{}, err
	}
	return a, nil
}
