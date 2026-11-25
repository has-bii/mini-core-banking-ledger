package auth

import (
	"backend/internal/modules/session"
	"backend/internal/modules/user"
	"context"
	"errors"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type userStore interface {
	Create(ctx context.Context, name, email, passwordHash string) (user.User, error)
	GetByID(ctx context.Context, userID string) (user.User, error)
	GetByEmail(ctx context.Context, email string) (user.User, error)
}

type sessionService interface {
	Create(ctx context.Context, userID string) (session.Session, string, error)
	Revoke(ctx context.Context, token string) error
}

type Service struct {
	users    userStore
	sessions sessionService
}

// NewService returns an authentication service backed by users and sessions.
func NewService(users userStore, sessions sessionService) *Service {
	return &Service{users: users, sessions: sessions}
}

// Register creates and returns a user.
// It returns an error matching user.ErrEmailTaken if the email is already registered.
// The caller must normalize and validate the DTO first.
func (s *Service) Register(ctx context.Context, payload RegisterDTO) (user.User, error) {
	_, err := s.users.GetByEmail(ctx, payload.Email)
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

	u, err := s.users.Create(ctx, payload.Name, payload.Email, hashedPassword)
	if errors.Is(err, user.ErrEmailTaken) {
		return user.User{}, err
	}
	if err != nil {
		return user.User{}, fmt.Errorf("register user: %w", err)
	}

	return u, nil
}

// A valid cost-10 hash keeps unknown-user checks comparable to password checks.
const dummyPasswordHash = "$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy"

// Login authenticates the user and returns the user, a new session, and its raw token.
// It returns an error matching ErrInvalidCredentials if the email is unknown or the password check fails.
// The caller must normalize and validate the DTO first.
func (s *Service) Login(ctx context.Context, payload LoginDTO) (user.User, session.Session, string, error) {
	u, err := s.users.GetByEmail(ctx, payload.Email)
	if errors.Is(err, user.ErrNotFound) {
		_ = comparePassword(dummyPasswordHash, payload.Password)
		return user.User{}, session.Session{}, "", errors.Join(ErrInvalidCredentials, err)
	}
	if err != nil {
		return user.User{}, session.Session{}, "", fmt.Errorf("find login user: %w", err)
	}

	if err := comparePassword(u.Password, payload.Password); err != nil {
		return user.User{}, session.Session{}, "", errors.Join(ErrInvalidCredentials, err)
	}

	created, token, err := s.sessions.Create(ctx, u.ID)
	if err != nil {
		return user.User{}, session.Session{}, "", fmt.Errorf("create login session: %w", err)
	}
	return u, created, token, nil
}

// Logout revokes the session identified by the raw token.
// It succeeds even if the token no longer identifies a session.
func (s *Service) Logout(ctx context.Context, token string) error {
	if err := s.sessions.Revoke(ctx, token); err != nil && !errors.Is(err, session.ErrNotFound) {
		return fmt.Errorf("revoke logout session: %w", err)
	}
	return nil
}

// Me returns the current user for a verified user ID.
// It returns an error matching ErrUnauthenticated if the user no longer exists.
func (s *Service) Me(ctx context.Context, userID string) (user.User, error) {
	u, err := s.users.GetByID(ctx, userID)
	if errors.Is(err, user.ErrNotFound) {
		return user.User{}, errors.Join(ErrUnauthenticated, err)
	}
	if err != nil {
		return user.User{}, fmt.Errorf("find current user: %w", err)
	}
	return u, nil
}

func comparePassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

func hashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(hash), err
}
