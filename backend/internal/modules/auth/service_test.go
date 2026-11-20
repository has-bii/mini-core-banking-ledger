package auth_test

import (
	"backend/internal/modules/auth"
	"backend/internal/modules/session"
	"backend/internal/modules/user"
	"context"
	"errors"
	"fmt"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

type userStoreStub struct {
	create     func(context.Context, string, string, string) (user.User, error)
	getByID    func(context.Context, string) (user.User, error)
	getByEmail func(context.Context, string) (user.User, error)
}

func (s userStoreStub) Create(ctx context.Context, name, email, passwordHash string) (user.User, error) {
	return s.create(ctx, name, email, passwordHash)
}

func (s userStoreStub) GetByID(ctx context.Context, userID string) (user.User, error) {
	return s.getByID(ctx, userID)
}

func (s userStoreStub) GetByEmail(ctx context.Context, email string) (user.User, error) {
	return s.getByEmail(ctx, email)
}

type sessionServiceStub struct {
	create func(context.Context, string) (session.Session, string, error)
	revoke func(context.Context, string) error
}

func (s sessionServiceStub) Create(ctx context.Context, userID string) (session.Session, string, error) {
	return s.create(ctx, userID)
}

func (s sessionServiceStub) Revoke(ctx context.Context, token string) error {
	return s.revoke(ctx, token)
}

func TestService_Register(t *testing.T) {
	t.Parallel()
	payload := auth.RegisterDTO{Name: "Alice", Email: "alice@example.com", Password: " password123 "}
	registered := user.User{ID: "user-id", Name: payload.Name, Email: payload.Email}
	storeErr := errors.New("user store unavailable")
	conflict := fmt.Errorf("insert user: %w", user.ErrEmailTaken)

	for _, tt := range []struct {
		name                          string
		lookupErr, createErr, wantErr error
	}{
		{name: "success", lookupErr: user.ErrNotFound},
		{name: "existing email", wantErr: user.ErrEmailTaken},
		{name: "lookup failure", lookupErr: storeErr, wantErr: storeErr},
		{name: "concurrent email conflict", lookupErr: user.ErrNotFound, createErr: conflict, wantErr: conflict},
		{name: "create failure", lookupErr: user.ErrNotFound, createErr: storeErr, wantErr: storeErr},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			users := userStoreStub{
				getByEmail: func(gotCtx context.Context, email string) (user.User, error) {
					if gotCtx != ctx || email != payload.Email {
						t.Error("GetByEmail() did not receive the caller context and registration email")
					}
					if tt.lookupErr != nil {
						return user.User{}, tt.lookupErr
					}
					return registered, nil
				},
				create: func(gotCtx context.Context, name, email, passwordHash string) (user.User, error) {
					if !errors.Is(tt.lookupErr, user.ErrNotFound) {
						t.Fatal("Create() called without an available email")
					}
					if gotCtx != ctx || name != payload.Name || email != payload.Email {
						t.Error("Create() did not receive the caller context, name, and email")
					}
					if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(payload.Password)); err != nil {
						t.Fatal("Create() did not receive a bcrypt hash of the original password")
					}
					return registered, tt.createErr
				},
			}

			got, err := auth.NewService(users, nil).Register(ctx, payload)
			want := registered
			if tt.wantErr != nil {
				want = user.User{}
			}
			if got != want || !errors.Is(err, tt.wantErr) || errors.Is(err, user.ErrEmailTaken) != errors.Is(tt.wantErr, user.ErrEmailTaken) {
				t.Errorf("Register() returned an unexpected user or error: %v (want %v)", err, tt.wantErr)
			}
		})
	}
}

func TestService_Login(t *testing.T) {
	t.Parallel()
	payload := auth.LoginDTO{Email: "alice@example.com", Password: "password123"}
	hash, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	account := user.User{ID: "user-id", Email: payload.Email, Password: string(hash)}
	created := session.Session{ID: "session-id", UserID: account.ID}
	const token = "session-token"
	lookupFailure := errors.New("user store unavailable")
	createFailure := errors.New("session store unavailable")

	for _, tt := range []struct {
		name                          string
		wrongPassword, corruptHash    bool
		invalidCredentials            bool
		lookupErr, createErr, wantErr error
	}{
		{name: "success"},
		{name: "unknown email", lookupErr: user.ErrNotFound, invalidCredentials: true, wantErr: user.ErrNotFound},
		{name: "wrong password", wrongPassword: true, invalidCredentials: true, wantErr: bcrypt.ErrMismatchedHashAndPassword},
		{name: "corrupt hash", corruptHash: true, invalidCredentials: true, wantErr: bcrypt.ErrHashTooShort},
		{name: "lookup failure", lookupErr: lookupFailure, wantErr: lookupFailure},
		{name: "session creation failure", createErr: createFailure, wantErr: createFailure},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			input, u := payload, account
			if tt.wrongPassword {
				input.Password = "different-password"
			}
			if tt.corruptHash {
				u.Password = "not-a-bcrypt-hash"
			}
			users := userStoreStub{getByEmail: func(gotCtx context.Context, email string) (user.User, error) {
				if gotCtx != ctx || email != input.Email {
					t.Error("GetByEmail() did not receive the caller context and login email")
				}
				return u, tt.lookupErr
			}}
			sessions := sessionServiceStub{create: func(gotCtx context.Context, userID string) (session.Session, string, error) {
				if tt.invalidCredentials || tt.lookupErr != nil {
					t.Fatal("Create() called before credentials were accepted")
				}
				if gotCtx != ctx || userID != u.ID {
					t.Error("Create() did not receive the caller context and logged-in user ID")
				}
				return created, token, tt.createErr
			}}

			got, gotSession, gotToken, err := auth.NewService(users, sessions).Login(ctx, input)
			if !errors.Is(err, tt.wantErr) || errors.Is(err, auth.ErrInvalidCredentials) != tt.invalidCredentials {
				t.Errorf("Login() error = %v; want cause %v, invalid credentials = %t", err, tt.wantErr, tt.invalidCredentials)
			}
			if tt.wantErr != nil {
				if got != (user.User{}) || gotSession != (session.Session{}) || gotToken != "" {
					t.Error("failed Login() returned a user, session, or token")
				}
			} else if got != u || gotSession != created || gotToken != token {
				t.Error("Login() did not return the user, session, and token")
			}
		})
	}
}

func TestService_Logout(t *testing.T) {
	t.Parallel()
	const token = "session-token"
	failure := errors.New("session store unavailable")
	for _, tt := range []struct {
		name           string
		cause, wantErr error
	}{
		{name: "revoked"},
		{name: "already absent", cause: fmt.Errorf("revoke session: %w", session.ErrNotFound)},
		{name: "revocation failure", cause: failure, wantErr: failure},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			called := false
			sessions := sessionServiceStub{revoke: func(gotCtx context.Context, gotToken string) error {
				called = true
				if gotCtx != ctx || gotToken != token {
					t.Error("Revoke() did not receive the caller context and token")
				}
				return tt.cause
			}}
			if err := auth.NewService(nil, sessions).Logout(ctx, token); !errors.Is(err, tt.wantErr) || !called {
				t.Errorf("Logout() error = %v, called Revoke() = %t; want %v and a revocation attempt", err, called, tt.wantErr)
			}
		})
	}
}

func TestService_Me(t *testing.T) {
	t.Parallel()
	account := user.User{ID: "user-id", Email: "alice@example.com"}
	for _, tt := range []struct {
		name, userID    string
		cause           error
		unauthenticated bool
	}{
		{name: "current user", userID: account.ID},
		{name: "missing user", userID: account.ID, cause: user.ErrNotFound, unauthenticated: true},
		{name: "invalid user ID", userID: "not-a-uuid", cause: user.ErrInvalidUserID},
		{name: "lookup failure", userID: account.ID, cause: errors.New("user store unavailable")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			users := userStoreStub{getByID: func(gotCtx context.Context, userID string) (user.User, error) {
				if gotCtx != ctx || userID != tt.userID {
					t.Error("GetByID() did not receive the caller context and user ID")
				}
				return account, tt.cause
			}}
			got, err := auth.NewService(users, nil).Me(ctx, tt.userID)
			want := account
			if tt.cause != nil {
				want = user.User{}
			}
			if got != want || !errors.Is(err, tt.cause) || errors.Is(err, auth.ErrUnauthenticated) != tt.unauthenticated {
				t.Errorf("Me() returned an unexpected user or error: %v (want cause %v, unauthenticated = %t)", err, tt.cause, tt.unauthenticated)
			}
		})
	}
}
