package auth

import (
	"backend/internal/modules/account"
	"backend/internal/modules/session"
	"backend/internal/modules/user"
	"backend/internal/platform/database"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/crypto/bcrypt"
)

type userStore interface {
	FindByID(context.Context, string) (user.User, error)
	FindByEmail(context.Context, string) (user.User, error)
}

type accountStore interface {
	FindByUserID(context.Context, string) (account.Account, error)
}

type sessionService interface {
	Create(context.Context, string) (session.Session, string, error)
	Revoke(context.Context, string) error
}

type Service struct {
	pool     *pgxpool.Pool
	users    userStore
	accounts accountStore
	sessions sessionService
}

func NewService(pool *pgxpool.Pool, users userStore, accounts accountStore, sessions sessionService) *Service {
	return &Service{pool: pool, users: users, accounts: accounts, sessions: sessions}
}

func (s *Service) Register(ctx context.Context, payload RegisterDTO) (user.User, error) {
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

// A valid cost-10 hash keeps unknown-user checks comparable to password checks.
const dummyPasswordHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

func (s *Service) Login(ctx context.Context, payload LoginDTO) (user.User, session.Session, string, error) {
	u, err := s.users.FindByEmail(ctx, payload.Email)
	if errors.Is(err, user.ErrNotFound) {
		_ = comparePassword(dummyPasswordHash, payload.Password)
		return user.User{}, session.Session{}, "", ErrInvalidCredentials
	}
	if err != nil {
		return user.User{}, session.Session{}, "", fmt.Errorf("find login user: %w", err)
	}

	a, err := s.accounts.FindByUserID(ctx, u.ID)
	if errors.Is(err, account.ErrNotFound) {
		_ = comparePassword(dummyPasswordHash, payload.Password)
		return user.User{}, session.Session{}, "", ErrInvalidCredentials
	}
	if err != nil {
		return user.User{}, session.Session{}, "", fmt.Errorf("find login account: %w", err)
	}
	if a.Password == nil {
		_ = comparePassword(dummyPasswordHash, payload.Password)
		return user.User{}, session.Session{}, "", ErrInvalidCredentials
	}
	if err := comparePassword(*a.Password, payload.Password); err != nil {
		return user.User{}, session.Session{}, "", ErrInvalidCredentials
	}

	created, token, err := s.sessions.Create(ctx, u.ID)
	if err != nil {
		return user.User{}, session.Session{}, "", fmt.Errorf("create login session: %w", err)
	}
	return u, created, token, nil
}

func (s *Service) Logout(ctx context.Context, token string) error {
	if err := s.sessions.Revoke(ctx, token); err != nil && !errors.Is(err, session.ErrNotFound) {
		return fmt.Errorf("revoke logout session: %w", err)
	}
	return nil
}

// Me returns the current user for a verified user ID.
func (s *Service) Me(ctx context.Context, userID string) (user.User, error) {
	u, err := s.users.FindByID(ctx, userID)
	if errors.Is(err, user.ErrNotFound) {
		return user.User{}, ErrUnauthenticated
	}
	if err != nil {
		return user.User{}, fmt.Errorf("find current user: %w", err)
	}
	return u, nil
}

func comparePassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func hashPassword(v string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(v), bcrypt.DefaultCost)
	return string(hash), err
}
