package wallet

import "context"

type walletStore interface {
	ListByUserID(ctx context.Context, userID string) ([]Wallet, error)
	GetByID(ctx context.Context, walletID, userID string) (Wallet, error)
	Create(ctx context.Context, userID string, data CreateWalletDTO) (Wallet, error)
}

type Service struct{ walletStore walletStore }

func NewService(walletStore walletStore) *Service { return &Service{walletStore: walletStore} }

func (s *Service) GetUserWallets(ctx context.Context, userID string) ([]Wallet, error) {
	return s.walletStore.ListByUserID(ctx, userID)
}

func (s *Service) GetUserWallet(ctx context.Context, walletID, userID string) (Wallet, error) {
	return s.walletStore.GetByID(ctx, walletID, userID)
}

// CreateUserWallet requires a DTO already normalized and validated by the caller.
func (s *Service) CreateUserWallet(ctx context.Context, userID string, data CreateWalletDTO) (Wallet, error) {
	return s.walletStore.Create(ctx, userID, data)
}
