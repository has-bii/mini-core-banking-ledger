package account

import (
	"context"
)

type accountStore interface {
	ListByUserID(ctx context.Context, userID string) ([]Account, error)
}

type Service struct{ accountStore accountStore }

func NewService(accountStore accountStore) *Service { return &Service{accountStore: accountStore} }

func (s *Service) GetUserAccounts(ctx context.Context, userID string) ([]Account, error) {
	return s.accountStore.ListByUserID(ctx, userID)
}
