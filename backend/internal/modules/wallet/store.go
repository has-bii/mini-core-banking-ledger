package wallet

import (
	"backend/internal/platform/postgres"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Store struct{ db postgres.Beginner }

// NewStore returns a wallet store backed by db, which must support transactions.
func NewStore(db postgres.Beginner) *Store { return &Store{db: db} }

type scanner interface{ Scan(dest ...any) error }

func scanWallet(r scanner) (Wallet, error) {
	var a Wallet
	err := r.Scan(&a.ID, &a.UserID, &a.Name, &a.Type, &a.SortOrder, &a.Balance, &a.CreatedAt)
	return a, err
}

// ListByUserID returns the user's wallets in ascending sort order. It wraps
// ErrInvalidUserID if userID is not a valid UUID.
//
// For a valid user ID with no matching wallets, the successful result is:
//
//	[]Wallet{}
func (s *Store) ListByUserID(ctx context.Context, userID string) ([]Wallet, error) {
	if err := uuid.Validate(userID); err != nil {
		return nil, fmt.Errorf("list by user id: %w", ErrInvalidUserID)
	}

	const q = `SELECT id, user_id, name, type, sort_order, balance, created_at FROM wallets
	WHERE user_id = $1
	ORDER BY sort_order ASC`

	rows, err := s.db.Query(ctx, q, userID)
	if err != nil {
		return nil, fmt.Errorf("list by user id: %w", err)
	}
	return pgx.CollectRows(rows, func(r pgx.CollectableRow) (Wallet, error) {
		return scanWallet(r)
	})
}

// GetByID returns the wallet with ID owned by userID. It returns ErrNotFound
// if the wallet does not exist or belongs to another user. Invalid UUIDs yield
// errors wrapping ErrInvalidWalletID or ErrInvalidUserID, respectively.
func (s *Store) GetByID(ctx context.Context, ID, userID string) (Wallet, error) {
	if err := uuid.Validate(ID); err != nil {
		return Wallet{}, fmt.Errorf("wallet by id: %w", ErrInvalidWalletID)
	}
	if err := uuid.Validate(userID); err != nil {
		return Wallet{}, fmt.Errorf("wallet by id: %w", ErrInvalidUserID)
	}

	const q = `SELECT id, user_id, name, type, sort_order, balance, created_at FROM wallets
	WHERE id = $1 AND user_id = $2`

	a, err := scanWallet(s.db.QueryRow(ctx, q, ID, userID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Wallet{}, ErrNotFound
	}
	if err != nil {
		return Wallet{}, fmt.Errorf("wallet by id: %w", err)
	}
	return a, nil
}

// Create inserts and returns a wallet after the user's existing wallets in sort
// order. It requires a DTO already normalized and validated by the caller.
// It returns ErrNotFound if the user is missing or has no existing wallets and
// wraps ErrInvalidUserID if userID is not a valid UUID.
func (s *Store) Create(ctx context.Context, userID string, data CreateWalletDTO) (Wallet, error) {
	if err := uuid.Validate(userID); err != nil {
		return Wallet{}, fmt.Errorf("create wallet: %w", ErrInvalidUserID)
	}

	var a Wallet
	err := postgres.InTx(ctx, s.db, func(tx pgx.Tx) error {
		// Serialize sort-order allocation for this user until the transaction commits.
		var lockedUserID string
		err := tx.QueryRow(ctx, `SELECT id FROM users WHERE id = $1 FOR UPDATE`, userID).Scan(&lockedUserID)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return fmt.Errorf("lock wallet owner: %w", err)
		}

		var lastSortOrder int
		err = tx.QueryRow(ctx,
			`SELECT sort_order FROM wallets WHERE user_id = $1 ORDER BY sort_order DESC LIMIT 1`,
			userID).Scan(&lastSortOrder)
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		if err != nil {
			return err
		}
		err = tx.QueryRow(ctx,
			`INSERT INTO wallets (user_id, name, type, sort_order)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, name, type, sort_order, balance, created_at`,
			userID, data.Name, data.WalletType, lastSortOrder+1,
		).Scan(&a.ID, &a.UserID, &a.Name, &a.Type, &a.SortOrder, &a.Balance, &a.CreatedAt)

		if err != nil {
			return fmt.Errorf("create wallet: %w", err)
		}
		return nil
	})
	if err != nil {
		return Wallet{}, err
	}
	return a, nil
}
