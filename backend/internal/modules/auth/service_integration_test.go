//go:build integration

package auth_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"backend/internal/modules/account"
	"backend/internal/modules/auth"
	"backend/internal/modules/user"
	"backend/internal/testutil/testdb"

	"github.com/jackc/pgx/v5/pgconn"
)

func TestRegisterRollsBackWhenCredentialsFail(t *testing.T) {
	t.Parallel()
	pool := testdb.Open(t)
	users := user.NewStore(pool)
	svc := auth.NewService(pool, users, account.NewStore(pool), nil)
	// Fault injection is restricted to this test's isolated schema.
	if _, err := pool.Exec(t.Context(), `ALTER TABLE accounts ADD CONSTRAINT reject_credentials CHECK (false)`); err != nil {
		t.Fatal(err)
	}
	payload := auth.RegisterDTO{Name: "Alice", Email: "alice@example.com", Password: "password123"}
	got, err := svc.Register(t.Context(), payload)
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23514" || got != (user.User{}) {
		t.Fatalf("Register() = %+v, %v; want empty user and credentials constraint failure", got, err)
	}
	if _, err := users.FindByEmail(t.Context(), payload.Email); !errors.Is(err, user.ErrNotFound) {
		t.Fatalf("failed registration left a user behind: %v", err)
	}
	if _, err := pool.Exec(t.Context(), `ALTER TABLE accounts DROP CONSTRAINT reject_credentials`); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Register(t.Context(), payload); err != nil {
		t.Fatalf("registration cannot retry after rollback: %v", err)
	}
}

// Both real reads finish before either registration can write. The test therefore
// exercises the database's unique constraint, not just the email pre-check.
type simultaneousReads struct {
	*user.Store
	ready sync.WaitGroup
	start chan struct{}
}

func (s *simultaneousReads) FindByEmail(ctx context.Context, email string) (user.User, error) {
	u, err := s.Store.FindByEmail(ctx, email)
	s.ready.Done()
	select {
	case <-s.start:
		return u, err
	case <-ctx.Done():
		return user.User{}, ctx.Err()
	}
}

func TestConcurrentRegistrationHasOneWinner(t *testing.T) {
	t.Parallel()
	pool := testdb.Open(t)
	users := &simultaneousReads{Store: user.NewStore(pool), start: make(chan struct{})}
	users.ready.Add(2)
	svc := auth.NewService(pool, users, account.NewStore(pool), nil)
	ctx, cancel := context.WithTimeout(t.Context(), 10*time.Second)
	defer cancel()
	type result struct {
		user user.User
		err  error
	}
	results := make(chan result, 2)
	for range 2 {
		go func() {
			u, err := svc.Register(ctx, auth.RegisterDTO{Name: "Alice", Email: "alice@example.com", Password: "password123"})
			results <- result{u, err}
		}()
	}
	go func() {
		users.ready.Wait()
		close(users.start)
	}()
	var winner user.User
	wins, conflicts := 0, 0
	for range 2 {
		select {
		case r := <-results:
			switch {
			case r.err == nil:
				winner = r.user
				wins++
			case errors.Is(r.err, user.ErrEmailTaken):
				conflicts++
				if r.user != (user.User{}) {
					t.Error("losing registration returned a user")
				}
			default:
				t.Errorf("unexpected registration error: %v", r.err)
			}
		case <-ctx.Done():
			t.Fatal("concurrent registrations did not finish")
		}
	}
	if wins != 1 || conflicts != 1 {
		t.Fatalf("wins = %d, conflicts = %d; want one of each", wins, conflicts)
	}
	if _, err := account.NewStore(pool).FindByUserID(t.Context(), winner.ID); err != nil {
		t.Fatalf("winning registration has no credentials: %v", err)
	}
}
