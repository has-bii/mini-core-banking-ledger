//go:build integration

package auth_test

import (
	"backend/internal/modules/account"
	"backend/internal/modules/auth"
	"backend/internal/modules/session"
	"backend/internal/modules/user"
	"backend/internal/testutil"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

func TestService_RegisterRollsBackCredentialsFailure(t *testing.T) {
	t.Parallel()
	pool := testutil.Postgres(t)
	users := user.NewStore(pool)
	svc := auth.NewService(pool, users, account.NewStore(pool), session.NewService(session.NewStore(pool)))
	if _, err := pool.Exec(t.Context(), "ALTER TABLE accounts ADD CONSTRAINT test_reject_credentials CHECK (false)"); err != nil {
		t.Fatal(err)
	}
	payload := auth.RegisterDTO{Name: "Alice", Email: "alice@example.com", Password: "password123"}
	_, err := svc.Register(t.Context(), payload)
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23514" || pgErr.ConstraintName != "test_reject_credentials" {
		t.Fatalf("Register() error = %v, want forced credentials constraint failure", err)
	}
	if _, err := users.FindByEmail(t.Context(), payload.Email); !errors.Is(err, user.ErrNotFound) {
		t.Fatalf("registration left a user after credentials failure: %v", err)
	}
	if _, err := pool.Exec(t.Context(), "ALTER TABLE accounts DROP CONSTRAINT test_reject_credentials"); err != nil {
		t.Fatal(err)
	}
	u, err := svc.Register(t.Context(), payload)
	if err != nil {
		t.Fatalf("registration after rollback must succeed: %v", err)
	}
	if _, err := account.NewStore(pool).FindByUserID(t.Context(), u.ID); err != nil {
		t.Fatalf("successful registration must create credentials: %v", err)
	}
}

func TestService_DependencyFailures(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, table string
	}{
		{"register lookup", "users"},
		{"login user lookup", "users"},
		{"login account lookup", "accounts"},
		{"login session creation", "sessions"},
		{"logout deletion", "sessions"},
		{"current user lookup", "users"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			pool := testutil.Postgres(t)
			svc := auth.NewService(pool, user.NewStore(pool), account.NewStore(pool), session.NewService(session.NewStore(pool)))
			payload := auth.RegisterDTO{Name: "Alice", Email: "alice@example.com", Password: "password123"}
			u, err := svc.Register(t.Context(), payload)
			if err != nil {
				t.Fatal(err)
			}
			// Only this test's schema is affected; each missing table targets one dependency.
			if _, err := pool.Exec(t.Context(), "DROP TABLE "+tt.table+" CASCADE"); err != nil {
				t.Fatal(err)
			}
			switch tt.name {
			case "register lookup":
				var got user.User
				got, err = svc.Register(t.Context(), payload)
				if got != (user.User{}) {
					t.Fatal("failed registration exposed a user")
				}
			case "current user lookup":
				var got user.User
				got, err = svc.Me(t.Context(), u.ID)
				if got != (user.User{}) {
					t.Fatal("failed current-user lookup exposed a user")
				}
			case "logout deletion":
				err = svc.Logout(t.Context(), "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA")
			default:
				var got user.User
				var created session.Session
				var token string
				got, created, token, err = svc.Login(t.Context(), auth.LoginDTO{Email: payload.Email, Password: payload.Password})
				if got != (user.User{}) || created != (session.Session{}) || token != "" {
					t.Fatal("failed login exposed a user, session, or token")
				}
			}
			var pgErr *pgconn.PgError
			if !errors.As(err, &pgErr) || pgErr.Code != "42P01" || !errors.Is(err, pgErr) {
				t.Fatalf("service error = %v; want wrapped missing-table error", err)
			}
		})
	}
}

func TestService_RegistrationAndIdentity(t *testing.T) {
	t.Parallel()
	pool := testutil.Postgres(t)
	svc := auth.NewService(pool, user.NewStore(pool), account.NewStore(pool), session.NewService(session.NewStore(pool)))
	payload := auth.RegisterDTO{Name: "Alice", Email: "alice@example.com", Password: "password123"}
	u, err := svc.Register(t.Context(), payload)
	if err != nil {
		t.Fatal(err)
	}
	if u.ID == "" || u.Name != payload.Name || u.Email != payload.Email || u.CreatedAt.IsZero() || u.UpdatedAt.IsZero() {
		t.Fatalf("Register() = %+v", u)
	}
	credentials, err := account.NewStore(pool).FindByUserID(t.Context(), u.ID)
	if err != nil {
		t.Fatal(err)
	}
	if credentials.Password == nil || *credentials.Password == payload.Password {
		t.Fatal("password must be stored as a hash")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*credentials.Password), []byte(payload.Password)); err != nil {
		t.Fatalf("stored password is not a valid bcrypt hash: %v", err)
	}
	if got, err := svc.Register(t.Context(), payload); !errors.Is(err, user.ErrEmailTaken) || got != (user.User{}) {
		t.Fatalf("duplicate Register() = %+v, %v", got, err)
	}
	if got, err := svc.Me(t.Context(), u.ID); err != nil || got != u {
		t.Fatalf("Me() = %+v, %v; want %+v", got, err, u)
	}
	if _, err := pool.Exec(t.Context(), "DELETE FROM users WHERE id = $1", u.ID); err != nil {
		t.Fatal(err)
	}
	if got, err := svc.Me(t.Context(), u.ID); !errors.Is(err, auth.ErrUnauthenticated) || got != (user.User{}) {
		t.Fatalf("deleted user Me() = %+v, %v", got, err)
	}
}

func TestService_LoginOutcomes(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"valid", "unknown email", "missing account", "no password", "wrong password", "corrupt hash"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			pool := testutil.Postgres(t)
			sessions := session.NewService(session.NewStore(pool))
			svc := auth.NewService(pool, user.NewStore(pool), account.NewStore(pool), sessions)
			u, err := svc.Register(t.Context(), auth.RegisterDTO{Name: "Alice", Email: "alice@example.com", Password: "password123"})
			if err != nil {
				t.Fatal(err)
			}
			payload := auth.LoginDTO{Email: u.Email, Password: "password123"}
			switch name {
			case "unknown email":
				payload.Email = "missing@example.com"
			case "missing account":
				_, err = pool.Exec(t.Context(), `DELETE FROM accounts WHERE "userId" = $1`, u.ID)
			case "no password":
				_, err = pool.Exec(t.Context(), `UPDATE accounts SET password = NULL WHERE "userId" = $1`, u.ID)
			case "wrong password":
				payload.Password = "incorrect"
			case "corrupt hash":
				_, err = pool.Exec(t.Context(), `UPDATE accounts SET password = 'not-a-bcrypt-hash' WHERE "userId" = $1`, u.ID)
			}
			if err != nil {
				t.Fatal(err)
			}
			got, created, token, err := svc.Login(t.Context(), payload)
			if name != "valid" {
				if !errors.Is(err, auth.ErrInvalidCredentials) || got != (user.User{}) || created != (session.Session{}) || token != "" {
					t.Fatalf("failed Login() = %+v, %+v, %q, %v", got, created, token, err)
				}
				var count int
				if err := pool.QueryRow(t.Context(), "SELECT count(*) FROM sessions").Scan(&count); err != nil || count != 0 {
					t.Fatalf("failed login persisted sessions: count=%d, err=%v", count, err)
				}
				return
			}
			if err != nil || got != u || created.UserID != u.ID || token == "" {
				t.Fatalf("Login() = %+v, %+v, %q, %v", got, created, token, err)
			}
			if found, err := sessions.FindByToken(t.Context(), token); err != nil || found != created {
				t.Fatalf("login session = %+v, %v; want %+v", found, err, created)
			}
			if err := svc.Logout(t.Context(), token); err != nil {
				t.Fatal(err)
			}
			if _, err := sessions.FindByToken(t.Context(), token); !errors.Is(err, session.ErrNotFound) {
				t.Fatalf("logout did not revoke session: %v", err)
			}
			for _, absent := range []string{token, "malformed"} {
				if err := svc.Logout(t.Context(), absent); err != nil {
					t.Fatalf("logout of absent session: %v", err)
				}
			}
		})
	}
}
