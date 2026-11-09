package auth

import (
	"backend/internal/modules/account"
	"backend/internal/modules/user"
	"backend/internal/platform/database"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	pool  *pgxpool.Pool
	users *user.Store
}

func NewService(pool *pgxpool.Pool) *Service {
	return &Service{pool: pool, users: user.NewStore(pool)}
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

	var u user.User
	err = database.InTx(ctx, s.pool, func(tx pgx.Tx) error {
		var err error
		u, err = user.NewStore(tx).Create(ctx, payload.Name, payload.Email)
		if err != nil {
			return err
		}
		_, err = account.NewStore(tx).Create(ctx, u.ID, hashedPassword)
		return err
	})
	if err != nil {
		return user.User{}, fmt.Errorf("register user: %w", err)
	}

	return u, nil
}

func hashPassword(v string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(v), bcrypt.DefaultCost)
	return string(hash), err
}
