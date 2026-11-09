package auth

import (
	"backend/internal/modules/account"
	"backend/internal/modules/session"
	"backend/internal/modules/user"
	"context"
	"errors"
	"testing"
)

type loginUsers struct{ err error }

func (s loginUsers) FindByEmail(context.Context, string) (user.User, error) {
	return user.User{ID: "user-id", Email: "alice@example.com"}, s.err
}

type loginAccounts struct {
	password *string
	err      error
}

func (s loginAccounts) FindByUserID(context.Context, string) (account.Account, error) {
	return account.Account{Password: s.password}, s.err
}

type loginSessions struct {
	calls int
	err   error
}

func (s *loginSessions) Create(_ context.Context, userID string) (session.Session, string, error) {
	s.calls++
	return session.Session{UserID: userID}, "secret-token", s.err
}

func TestLoginWithoutPassword(t *testing.T) {
	sessions := &loginSessions{}
	svc := &Service{users: loginUsers{}, accounts: loginAccounts{}, sessions: sessions}
	_, _, token, err := svc.Login(context.Background(), LoginDTO{Email: "alice@example.com", Password: "password123"})
	if !errors.Is(err, ErrInvalidCredentials) || token != "" || sessions.calls != 0 {
		t.Fatalf("token = %q, error = %v, sessions = %d", token, err, sessions.calls)
	}
}

func TestLogin(t *testing.T) {
	hash, err := hashPassword("password123")
	if err != nil {
		t.Fatal(err)
	}
	failure := errors.New("database unavailable")
	for _, tc := range []struct {
		name, password                           string
		userErr, accountErr, sessionErr, wantErr error
		wantSession                              bool
	}{
		{name: "success", password: "password123", wantSession: true},
		{name: "wrong password", password: "incorrect", wantErr: ErrInvalidCredentials},
		{name: "unknown user", password: "password123", userErr: user.ErrNotFound, wantErr: ErrInvalidCredentials},
		{name: "missing account", password: "password123", accountErr: account.ErrNotFound, wantErr: ErrInvalidCredentials},
		{name: "user lookup failure", password: "password123", userErr: failure, wantErr: failure},
		{name: "account lookup failure", password: "password123", accountErr: failure, wantErr: failure},
		{name: "session failure", password: "password123", sessionErr: failure, wantErr: failure, wantSession: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sessions := &loginSessions{err: tc.sessionErr}
			svc := &Service{users: loginUsers{tc.userErr}, accounts: loginAccounts{&hash, tc.accountErr}, sessions: sessions}
			u, created, token, err := svc.Login(context.Background(), LoginDTO{Email: "alice@example.com", Password: tc.password})
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("error = %v, want %v", err, tc.wantErr)
			}
			if (sessions.calls == 1) != tc.wantSession {
				t.Fatalf("session calls = %d", sessions.calls)
			}
			if err != nil && token != "" {
				t.Fatal("returned token on failure")
			}
			if err == nil && (u.ID != "user-id" || created.UserID != u.ID || token != "secret-token") {
				t.Fatal("unexpected login result")
			}
		})
	}
}
