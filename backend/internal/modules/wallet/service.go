package wallet

import "context"

type walletStore interface {
	ListByUserID(ctx context.Context, userID string) ([]Wallet, error)
	GetByID(ctx context.Context, walletID, userID string) (Wallet, error)
	Create(ctx context.Context, userID string, data CreateWalletDTO) (Wallet, error)
}

type Service struct{ walletStore walletStore }

// NewService returns a wallet service backed by walletStore.
func NewService(walletStore walletStore) *Service { return &Service{walletStore: walletStore} }

// GetUserWallets returns the user's wallets in ascending sort order and passes
// through store errors unchanged.
func (s *Service) GetUserWallets(ctx context.Context, userID string) ([]Wallet, error) {
	return s.walletStore.ListByUserID(ctx, userID)
}

// GetUserWallet returns the wallet with walletID owned by userID and passes
// through store errors, including ErrNotFound if no wallet matches.
func (s *Service) GetUserWallet(ctx context.Context, walletID, userID string) (Wallet, error) {
	return s.walletStore.GetByID(ctx, walletID, userID)
}

// CreateUserWallet creates and returns a wallet for userID. It requires a DTO
// already normalized and validated by the caller and passes through store errors
// unchanged.
func (s *Service) CreateUserWallet(ctx context.Context, userID string, data CreateWalletDTO) (Wallet, error) {
	return s.walletStore.Create(ctx, userID, data)
}
