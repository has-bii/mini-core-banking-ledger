package wallet

import (
	"errors"
	"time"
)

var (
	ErrNotFound        = errors.New("wallet not found")
	ErrInvalidWalletID = errors.New("invalid wallet ID")
	ErrInvalidUserID   = errors.New("invalid user ID")
)

type WalletType string

const (
	Main      WalletType = "main"
	SubWallet WalletType = "sub_wallet"
	Savings   WalletType = "savings"
)

type Wallet struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Name      string     `json:"name"`
	Type      WalletType `json:"wallet_type"`
	SortOrder int16      `json:"sort_order"`
	Balance   int64      `json:"balance"`
	CreatedAt time.Time  `json:"created_at"`
}
