package auth

import (
	"backend/internal/modules/user"
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	users *user.Store
}

func NewService(users *user.Store) *Service {
	return &Service{users: users}
}

func (s *Service) Register(ctx context.Context, payload RegisterPayload) (user.User, error) {
	_, err := s.users.FindByEmail(ctx, payload.Email)
	if err == nil {
		return user.User{}, user.ErrEmailTaken
	}
	if !errors.Is(err, user.ErrNotFound) {
		return user.User{}, fmt.Errorf("check registration email: %w", err)
	}

	hashedPassword, err := hashPassword(payload.Password)
	if err != nil {
		return user.User{}, fmt.Errorf("hash password: %w", err)
	}

	u, err := s.users.CreateWithCredentials(ctx, payload.Name, payload.Email, hashedPassword)
	if err != nil {
		return user.User{}, fmt.Errorf("register user: %w", err)
	}

	return u, nil
}

func hashPassword(v string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(v), bcrypt.DefaultCost)
	return string(hash), err
}
