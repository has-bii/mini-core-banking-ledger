package auth_test

import (
	"backend/internal/modules/account"
	"backend/internal/modules/auth"
	"backend/internal/modules/session"
	"backend/internal/modules/user"
	"context"
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

type userStoreStub struct {
	result          user.User
	emailErr, idErr error
}

func (s userStoreStub) FindByID(context.Context, string) (user.User, error) {
	return s.result, s.idErr
}
func (s userStoreStub) FindByEmail(context.Context, string) (user.User, error) {
	return s.result, s.emailErr
}

type accountStoreStub struct {
	result account.Account
	err    error
}

func (s accountStoreStub) FindByUserID(context.Context, string) (account.Account, error) {
	return s.result, s.err
}

type sessionServiceStub struct {
	result               session.Session
	createErr, revokeErr error
}

func (s sessionServiceStub) Create(context.Context, string) (session.Session, string, error) {
	return s.result, "raw-token", s.createErr
}
func (s sessionServiceStub) FindByToken(context.Context, string) (session.Session, error) {
	return s.result, nil
}
func (s sessionServiceStub) Revoke(context.Context, string) error { return s.revokeErr }

func TestService_Register(t *testing.T) {
	t.Parallel()
	failure := errors.New("user lookup unavailable")
	for _, tt := range []struct {
		name    string
		readErr error
		wantErr error
	}{
		{"email already registered", nil, user.ErrEmailTaken},
		{"lookup failed", failure, failure},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := auth.NewService(nil, userStoreStub{emailErr: tt.readErr}, nil, nil)
			got, err := svc.Register(t.Context(), auth.RegisterDTO{Name: "Alice", Email: "alice@example.com", Password: "password123"})
			if !errors.Is(err, tt.wantErr) || got != (user.User{}) {
				t.Fatalf("Register() = %+v, %v; want no user and %v", got, err, tt.wantErr)
			}
		})
	}
}

func TestService_Login(t *testing.T) {
	t.Parallel()
	hash, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	passwordHash, corruptHash := string(hash), "not-a-bcrypt-hash"
	failure := errors.New("dependency unavailable")
	for _, tt := range []struct {
		name       string
		userErr    error
		accountErr error
		hash       *string
		password   string
		sessionErr error
		wantErr    error
	}{
		{name: "valid", hash: &passwordHash, password: "password123"},
		{name: "unknown email", userErr: user.ErrNotFound, wantErr: auth.ErrInvalidCredentials},
		{name: "user lookup failed", userErr: failure, wantErr: failure},
		{name: "missing account", accountErr: account.ErrNotFound, wantErr: auth.ErrInvalidCredentials},
		{name: "account lookup failed", accountErr: failure, wantErr: failure},
		{name: "no password", wantErr: auth.ErrInvalidCredentials},
		{name: "wrong password", hash: &passwordHash, password: "incorrect", wantErr: auth.ErrInvalidCredentials},
		{name: "corrupt hash", hash: &corruptHash, wantErr: auth.ErrInvalidCredentials},
		{name: "session creation failed", hash: &passwordHash, password: "password123", sessionErr: failure, wantErr: failure},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			u := user.User{ID: "user-id", Email: "alice@example.com"}
			current := session.Session{ID: "session-id", UserID: u.ID}
			svc := auth.NewService(nil,
				userStoreStub{result: u, emailErr: tt.userErr},
				accountStoreStub{result: account.Account{Password: tt.hash}, err: tt.accountErr},
				sessionServiceStub{result: current, createErr: tt.sessionErr},
			)
			got, created, token, err := svc.Login(t.Context(), auth.LoginDTO{Email: u.Email, Password: tt.password})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Login() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil {
				if got != (user.User{}) || created != (session.Session{}) || token != "" {
					t.Fatal("failed login exposed a user, session or token")
				}
			} else if got != u || created != current || token != "raw-token" {
				t.Fatalf("Login() = %+v, %+v, %q; want user and created session/token", got, created, token)
			}
		})
	}
}

func TestService_Logout(t *testing.T) {
	t.Parallel()
	failure := errors.New("session deletion unavailable")
	for _, tt := range []struct {
		name     string
		storeErr error
		wantErr  error
	}{
		{"revoked", nil, nil},
		{"already absent", session.ErrNotFound, nil},
		{"deletion failed", failure, failure},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := auth.NewService(nil, nil, nil, sessionServiceStub{revokeErr: tt.storeErr})
			if err := svc.Logout(t.Context(), "token"); !errors.Is(err, tt.wantErr) {
				t.Fatalf("Logout() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestService_Me(t *testing.T) {
	t.Parallel()
	failure := errors.New("dependency unavailable")
	for _, tt := range []struct {
		name    string
		userErr error
		wantErr error
	}{
		{"authenticated", nil, nil},
		{"deleted user", user.ErrNotFound, auth.ErrUnauthenticated},
		{"user lookup failed", failure, failure},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			u := user.User{ID: "user-id", Email: "alice@example.com"}
			svc := auth.NewService(nil, userStoreStub{result: u, idErr: tt.userErr}, nil, nil)
			got, err := svc.Me(t.Context(), u.ID)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Me() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr != nil && got != (user.User{}) {
				t.Fatal("failed authentication exposed a user")
			}
			if tt.wantErr == nil && got != u {
				t.Fatalf("Me() = %+v, want %+v", got, u)
			}
		})
	}
}
