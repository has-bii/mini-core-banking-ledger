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
