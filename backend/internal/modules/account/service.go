package account

import (
	"context"
)

type accountStore interface {
	ListByUserID(context.Context, string) ([]Account, error)
	GetByID(context.Context, string, string) (Account, error)
	Create(context.Context, string, *CreateAccountDTO) (Account, error)
}

type Service struct{ accountStore accountStore }

func NewService(accountStore accountStore) *Service { return &Service{accountStore: accountStore} }

func (s *Service) GetUserAccounts(ctx context.Context, userID string) ([]Account, error) {
	return s.accountStore.ListByUserID(ctx, userID)
}

func (s *Service) GetUserAccount(ctx context.Context, ID, userID string) (Account, error) {
	return s.accountStore.GetByID(ctx, ID, userID)
}

func (s *Service) CreateUserAccount(ctx context.Context, userID string, data *CreateAccountDTO) (Account, error) {
	return s.accountStore.Create(ctx, userID, data)
}
