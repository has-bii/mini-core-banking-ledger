package auth_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"backend/internal/modules/account"
	"backend/internal/modules/auth"
	"backend/internal/modules/session"
	"backend/internal/modules/user"

	"golang.org/x/crypto/bcrypt"
)

type fakeUsers struct {
	result    user.User
	err       error
	email, id string
}

func (f *fakeUsers) FindByEmail(_ context.Context, email string) (user.User, error) {
	f.email = email
	return f.result, f.err
}

func (f *fakeUsers) FindByID(_ context.Context, id string) (user.User, error) {
	f.id = id
	return f.result, f.err
}

type fakeAccounts struct {
	result account.Account
	err    error
	userID string
}

func (f *fakeAccounts) FindByUserID(_ context.Context, id string) (account.Account, error) {
	f.userID = id
	return f.result, f.err
}

type fakeSessions struct {
	result                               session.Session
	err                                  error
	createdFor, foundToken, revokedToken string
}

func (f *fakeSessions) Create(_ context.Context, id string) (session.Session, string, error) {
	f.createdFor = id
	return f.result, "raw-token", f.err
}

func (f *fakeSessions) FindByToken(_ context.Context, token string) (session.Session, error) {
	f.foundToken = token
	return f.result, f.err
}

func (f *fakeSessions) Revoke(_ context.Context, token string) error {
	f.revokedToken = token
	return f.err
}

func TestLogout(t *testing.T) {
	failure := errors.New("store unavailable")
	for _, tt := range []struct {
		name      string
		revokeErr error
		wantErr   error
	}{
		{name: "revoked"},
		{name: "not found", revokeErr: fmt.Errorf("revoke: %w", session.ErrNotFound)},
		{name: "store failure", revokeErr: failure, wantErr: failure},
	} {
		t.Run(tt.name, func(t *testing.T) {
			sessions := &fakeSessions{err: tt.revokeErr}
			svc := auth.NewService(nil, nil, nil, sessions)
			if err := svc.Logout(t.Context(), "raw-token"); !errors.Is(err, tt.wantErr) {
				t.Errorf("Logout() error = %v, want %v", err, tt.wantErr)
			}
			if sessions.revokedToken != "raw-token" {
				t.Errorf("revoked token = %q, want raw-token", sessions.revokedToken)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	hash, err := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	passwordHash := string(hash)
	failure := errors.New("store unavailable")
	wantUser := user.User{ID: "user-1", Email: "alice@example.com"}
	wantSession := session.Session{ID: "session-1", UserID: wantUser.ID}

	for _, tt := range []struct {
		name       string
		userErr    error
		accountErr error
		password   *string
		input      string
		sessionErr error
		wantErr    error
	}{
		{name: "success", password: &passwordHash, input: "password"},
		{name: "unknown user", userErr: fmt.Errorf("lookup: %w", user.ErrNotFound), wantErr: auth.ErrInvalidCredentials},
		{name: "user lookup failure", userErr: failure, wantErr: failure},
		{name: "missing account", accountErr: fmt.Errorf("lookup: %w", account.ErrNotFound), wantErr: auth.ErrInvalidCredentials},
		{name: "account lookup failure", accountErr: failure, wantErr: failure},
		{name: "no password", wantErr: auth.ErrInvalidCredentials},
		{name: "wrong password", password: &passwordHash, input: "wrong-password", wantErr: auth.ErrInvalidCredentials},
		{name: "session creation failure", password: &passwordHash, input: "password", sessionErr: failure, wantErr: failure},
	} {
		t.Run(tt.name, func(t *testing.T) {
			users := &fakeUsers{result: wantUser, err: tt.userErr}
			accounts := &fakeAccounts{result: account.Account{Password: tt.password}, err: tt.accountErr}
			sessions := &fakeSessions{result: wantSession, err: tt.sessionErr}
			svc := auth.NewService(nil, users, accounts, sessions)

			gotUser, gotSession, token, err := svc.Login(t.Context(), auth.LoginDTO{Email: wantUser.Email, Password: tt.input})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Login() error = %v, want %v", err, tt.wantErr)
			}
			if users.email != wantUser.Email {
				t.Errorf("email = %q, want %q", users.email, wantUser.Email)
			}
			wantAccountID := ""
			if tt.userErr == nil {
				wantAccountID = wantUser.ID
			}
			if accounts.userID != wantAccountID {
				t.Errorf("account lookup = %q, want %q", accounts.userID, wantAccountID)
			}
			wantSessionID := ""
			if tt.wantErr == nil || tt.sessionErr != nil {
				wantSessionID = wantUser.ID
			}
			if sessions.createdFor != wantSessionID {
				t.Errorf("session created for = %q, want %q", sessions.createdFor, wantSessionID)
			}
			if tt.wantErr == nil {
				if gotUser != wantUser || gotSession != wantSession || token != "raw-token" {
					t.Fatalf("unexpected login result: %v, %v, %q", gotUser, gotSession, token)
				}
			} else if gotUser != (user.User{}) || gotSession != (session.Session{}) || token != "" {
				t.Error("failed login returned user, session or token")
			}
		})
	}
}

func TestMe(t *testing.T) {
	failure := errors.New("store unavailable")
	want := user.User{ID: "user-1"}
	for _, tt := range []struct {
		name       string
		sessionErr error
		userErr    error
		wantErr    error
	}{
		{name: "success"},
		{name: "missing session", sessionErr: fmt.Errorf("lookup: %w", session.ErrNotFound), wantErr: auth.ErrUnauthenticated},
		{name: "session lookup failure", sessionErr: failure, wantErr: failure},
		{name: "missing user", userErr: fmt.Errorf("lookup: %w", user.ErrNotFound), wantErr: auth.ErrUnauthenticated},
		{name: "user lookup failure", userErr: failure, wantErr: failure},
	} {
		t.Run(tt.name, func(t *testing.T) {
			users := &fakeUsers{result: want, err: tt.userErr}
			sessions := &fakeSessions{result: session.Session{UserID: want.ID}, err: tt.sessionErr}
			svc := auth.NewService(nil, users, nil, sessions)

			got, err := svc.Me(t.Context(), "raw-token")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Me() error = %v, want %v", err, tt.wantErr)
			}
			if sessions.foundToken != "raw-token" {
				t.Errorf("session token = %q", sessions.foundToken)
			}
			wantUserID := ""
			if tt.sessionErr == nil {
				wantUserID = want.ID
			}
			if users.id != wantUserID {
				t.Errorf("user lookup = %q, want %q", users.id, wantUserID)
			}
			if tt.wantErr == nil && got != want || tt.wantErr != nil && got != (user.User{}) {
				t.Errorf("unexpected user: %v", got)
			}
		})
	}
}

func TestRegisterPreflight(t *testing.T) {
	failure := errors.New("store unavailable")
	for _, tt := range []struct {
		name    string
		findErr error
		wantErr error
	}{
		{"existing email", nil, user.ErrEmailTaken},
		{"lookup failure", failure, failure},
	} {
		t.Run(tt.name, func(t *testing.T) {
			users := &fakeUsers{err: tt.findErr}
			svc := auth.NewService(nil, users, nil, nil)
			got, err := svc.Register(t.Context(), auth.RegisterDTO{Email: "alice@example.com", Password: "password"})
			if !errors.Is(err, tt.wantErr) || got != (user.User{}) {
				t.Errorf("Register() = %v, %v; want empty user, %v", got, err, tt.wantErr)
			}
			if users.email != "alice@example.com" {
				t.Errorf("email = %q", users.email)
			}
		})
	}
}
