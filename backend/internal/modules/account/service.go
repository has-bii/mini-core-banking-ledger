package account

import (
	"backend/internal/platform/validation"
	"context"
)

type accountStore interface {
	FindAllByUserID(context.Context, string, validation.QueryOrder) ([]Account, error)
}

type Service struct{ store accountStore }

func NewService(store accountStore) *Service { return &Service{store: store} }

func (s *Service) GetAccountsByUserID(ctx context.Context, userID string, order validation.QueryOrder) ([]Account, error) {
	return s.store.FindAllByUserID(ctx, userID, order)
}
