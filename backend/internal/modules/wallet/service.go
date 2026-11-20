package wallet

import "context"

type walletStore interface {
	ListByUserID(context.Context, string) ([]Wallet, error)
	GetByID(context.Context, string, string) (Wallet, error)
	Create(context.Context, string, *CreateWalletDTO) (Wallet, error)
}

type Service struct{ walletStore walletStore }

func NewService(walletStore walletStore) *Service { return &Service{walletStore: walletStore} }

func (s *Service) GetUserWallets(ctx context.Context, userID string) ([]Wallet, error) {
	return s.walletStore.ListByUserID(ctx, userID)
}

func (s *Service) GetUserWallet(ctx context.Context, ID, userID string) (Wallet, error) {
	return s.walletStore.GetByID(ctx, ID, userID)
}

func (s *Service) CreateUserWallet(ctx context.Context, userID string, data *CreateWalletDTO) (Wallet, error) {
	return s.walletStore.Create(ctx, userID, data)
}
