package auth_test

import (
	"backend/internal/modules/auth"
	"backend/internal/modules/user"
	"backend/internal/platform/apperror"
	"context"
	"errors"
	"fmt"
	"testing"
)

// Failure injection: another registration wins after the email pre-check.
type registrationRaceStore struct{ cause error }

func (s registrationRaceStore) FindByEmail(context.Context, string) (user.User, error) {
	return user.User{}, user.ErrNotFound
}
func (s registrationRaceStore) Create(context.Context, string, string, string) (user.User, error) {
	return user.User{}, s.cause
}
func (s registrationRaceStore) FindByID(context.Context, string) (user.User, error) {
	return user.User{}, user.ErrNotFound
}

func TestService_RegisterPreservesConcurrentEmailConflict(t *testing.T) {
	t.Parallel()
	cause := fmt.Errorf("insert user: %w", user.ErrEmailTaken)
	svc := auth.NewService(registrationRaceStore{cause: cause}, nil)
	got, err := svc.Register(t.Context(), auth.RegisterDTO{Name: "Alice", Email: "alice@example.com", Password: "password123"})
	var appErr *apperror.Error
	if got != (user.User{}) || !errors.As(err, &appErr) || appErr.Code != "EMAIL_TAKEN" || appErr.Message != "email already taken" || !errors.Is(err, cause) {
		t.Fatalf("Register() = %+v, %v; want safe EMAIL_TAKEN preserving store cause", got, err)
	}
}
