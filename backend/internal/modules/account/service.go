package account

import (
	"backend/internal/platform/apperror"
	"backend/internal/platform/validation"
	"context"
	"errors"
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
	a, err := s.accountStore.GetByID(ctx, ID, userID)
	switch {
	case errors.Is(err, validation.ErrInvalidUUID):
		return Account{}, apperror.New(apperror.CodeValidationError, "invalid account ID", nil, err)
	case errors.Is(err, ErrNotFound):
		return Account{}, apperror.New(apperror.CodeNotFound, ErrNotFound.Error(), nil, err)
	default:
		return a, err
	}
}

func (s *Service) CreateUserAccount(ctx context.Context, userID string, data *CreateAccountDTO) (Account, error) {
	a, err := s.accountStore.Create(ctx, userID, data)
	var fields validation.Errors
	switch {
	case errors.As(err, &fields):
		return Account{}, apperror.New(apperror.CodeValidationError, "Some fields are invalid.", fields, err)
	case errors.Is(err, ErrNotFound):
		return Account{}, apperror.New(apperror.CodeNotFound, ErrNotFound.Error(), nil, err)
	default:
		return a, err
	}
}
