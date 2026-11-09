package auth

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"backend/internal/modules/account"
	"backend/internal/modules/session"
	"backend/internal/modules/user"

	"golang.org/x/crypto/bcrypt"
)

type fakeUsers struct {
	findByEmail func(context.Context, string) (user.User, error)
	findByID    func(context.Context, string) (user.User, error)
}

func (f fakeUsers) FindByEmail(ctx context.Context, email string) (user.User, error) {
	return f.findByEmail(ctx, email)
}

func (f fakeUsers) FindByID(ctx context.Context, id string) (user.User, error) {
	return f.findByID(ctx, id)
}

type fakeAccounts struct {
	find func(context.Context, string) (account.Account, error)
}

func (f fakeAccounts) FindByUserID(ctx context.Context, id string) (account.Account, error) {
	return f.find(ctx, id)
}

type fakeSessions struct {
	create func(context.Context, string) (session.Session, string, error)
	find   func(context.Context, string) (session.Session, error)
}

func (f fakeSessions) Create(ctx context.Context, id string) (session.Session, string, error) {
	return f.create(ctx, id)
}

func (f fakeSessions) FindByToken(ctx context.Context, token string) (session.Session, error) {
	return f.find(ctx, token)
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
		{name: "unknown user", userErr: fmt.Errorf("lookup: %w", user.ErrNotFound), wantErr: ErrInvalidCredentials},
		{name: "user lookup failure", userErr: failure, wantErr: failure},
		{name: "missing account", accountErr: fmt.Errorf("lookup: %w", account.ErrNotFound), wantErr: ErrInvalidCredentials},
		{name: "account lookup failure", accountErr: failure, wantErr: failure},
		{name: "no password", wantErr: ErrInvalidCredentials},
		{name: "wrong password", password: &passwordHash, input: "wrong-password", wantErr: ErrInvalidCredentials},
		{name: "session creation failure", password: &passwordHash, input: "password", sessionErr: failure, wantErr: failure},
	} {
		t.Run(tt.name, func(t *testing.T) {
			created := false
			service := NewService(nil, fakeUsers{findByEmail: func(_ context.Context, email string) (user.User, error) {
				if email != wantUser.Email {
					t.Errorf("email = %q, want %q", email, wantUser.Email)
				}
				return wantUser, tt.userErr
			}}, fakeAccounts{find: func(_ context.Context, id string) (account.Account, error) {
				if tt.userErr != nil {
					t.Fatal("account lookup after failed user lookup")
				}
				if id != wantUser.ID {
					t.Errorf("account user ID = %q", id)
				}
				return account.Account{Password: tt.password}, tt.accountErr
			}}, fakeSessions{create: func(_ context.Context, id string) (session.Session, string, error) {
				created = true
				if id != wantUser.ID {
					t.Errorf("session user ID = %q", id)
				}
				return wantSession, "raw-token", tt.sessionErr
			}})
			gotUser, gotSession, token, err := service.Login(context.Background(), LoginDTO{Email: wantUser.Email, Password: tt.input})
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Login() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil {
				if gotUser != wantUser || gotSession != wantSession || token != "raw-token" || !created {
					t.Fatalf("unexpected login result: %v, %v, %q", gotUser, gotSession, token)
				}
			} else {
				if gotUser != (user.User{}) || gotSession != (session.Session{}) || token != "" {
					t.Error("failed login returned user, session or token")
				}
				if created != (tt.sessionErr != nil) {
					t.Error("session creation must only follow successful authentication")
				}
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
		{name: "missing session", sessionErr: fmt.Errorf("lookup: %w", session.ErrNotFound), wantErr: ErrUnauthenticated},
		{name: "session lookup failure", sessionErr: failure, wantErr: failure},
		{name: "missing user", userErr: fmt.Errorf("lookup: %w", user.ErrNotFound), wantErr: ErrUnauthenticated},
		{name: "user lookup failure", userErr: failure, wantErr: failure},
	} {
		t.Run(tt.name, func(t *testing.T) {
			service := NewService(nil, fakeUsers{findByID: func(_ context.Context, id string) (user.User, error) {
				if tt.sessionErr != nil {
					t.Fatal("user lookup after failed session lookup")
				}
				if id != want.ID {
					t.Errorf("user ID = %q", id)
				}
				return want, tt.userErr
			}}, nil, fakeSessions{find: func(_ context.Context, token string) (session.Session, error) {
				if token != "raw-token" {
					t.Errorf("token = %q", token)
				}
				return session.Session{UserID: want.ID}, tt.sessionErr
			}})
			got, err := service.Me(context.Background(), "raw-token")
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("Me() error = %v, want %v", err, tt.wantErr)
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
			service := NewService(nil, fakeUsers{findByEmail: func(_ context.Context, email string) (user.User, error) {
				if email != "alice@example.com" {
					t.Errorf("email = %q", email)
				}
				return user.User{}, tt.findErr
			}}, nil, nil)
			got, err := service.Register(context.Background(), RegisterDTO{Email: "alice@example.com", Password: "password"})
			if !errors.Is(err, tt.wantErr) || got != (user.User{}) {
				t.Errorf("Register() = %v, %v; want empty user, %v", got, err, tt.wantErr)
			}
		})
	}
}
