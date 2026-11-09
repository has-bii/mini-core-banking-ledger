//go:build integration

package session_test

import (
	"errors"
	"testing"
	"time"

	"backend/internal/modules/session"
	"backend/internal/modules/user"
	"backend/internal/testutil/testdb"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestStoreSessionLifetimeAndRevocation(t *testing.T) {
	t.Parallel()
	pool := testdb.Open(t)
	u, err := user.NewStore(pool).Create(t.Context(), "Alice", "alice@example.com")
	if err != nil {
		t.Fatal(err)
	}
	store := session.NewStore(pool)
	expires := time.Now().UTC().Truncate(time.Second).Add(time.Hour)
	created, err := store.Create(t.Context(), u.ID, "active-hash", expires)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := uuid.Parse(created.ID); err != nil || created.UserID != u.ID || !created.ExpiresAt.Equal(expires) {
		t.Fatalf("unexpected session: %+v", created)
	}
	found, err := store.FindByTokenHash(t.Context(), "active-hash")
	if err != nil || found != created {
		t.Fatalf("session did not round-trip: %+v, %v", found, err)
	}
	if _, err := store.Create(t.Context(), u.ID, "expired-hash", time.Now().Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
	for _, hash := range []string{"expired-hash", "missing-hash"} {
		if _, err := store.FindByTokenHash(t.Context(), hash); !errors.Is(err, session.ErrNotFound) {
			t.Errorf("lookup %q error = %v, want ErrNotFound", hash, err)
		}
	}
	if _, err := store.Create(t.Context(), u.ID, "other-hash", expires); err != nil {
		t.Fatal(err)
	}
	for range 2 {
		if err := store.DeleteByTokenHash(t.Context(), "active-hash"); err != nil {
			t.Fatalf("revocation must be idempotent: %v", err)
		}
	}
	if _, err := store.FindByTokenHash(t.Context(), "active-hash"); !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("revoked session can still be found: %v", err)
	}
	if _, err := store.FindByTokenHash(t.Context(), "other-hash"); err != nil {
		t.Fatalf("revocation affected a different session: %v", err)
	}
}

func TestStoreSessionConstraints(t *testing.T) {
	t.Parallel()
	pool := testdb.Open(t)
	u, err := user.NewStore(pool).Create(t.Context(), "Alice", "alice@example.com")
	if err != nil {
		t.Fatal(err)
	}
	store := session.NewStore(pool)
	expires := time.Now().Add(time.Hour)
	if _, err := store.Create(t.Context(), u.ID, "same-hash", expires); err != nil {
		t.Fatal(err)
	}
	_, err = store.Create(t.Context(), u.ID, "same-hash", expires)
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "23505" {
		t.Errorf("duplicate token hash must preserve unique violation: %v", err)
	}
	_, err = store.Create(t.Context(), uuid.NewString(), "other-hash", expires)
	if !errors.As(err, &pgErr) || pgErr.Code != "23503" {
		t.Errorf("session with no user must preserve foreign-key violation: %v", err)
	}
}

func TestSessionServicePersistsHashedTokens(t *testing.T) {
	t.Parallel()
	pool := testdb.Open(t)
	u, err := user.NewStore(pool).Create(t.Context(), "Alice", "alice@example.com")
	if err != nil {
		t.Fatal(err)
	}
	store := session.NewStore(pool)
	svc := session.NewService(store)
	created, token, err := svc.Create(t.Context(), u.ID)
	if err != nil {
		t.Fatal(err)
	}
	found, err := svc.FindByToken(t.Context(), token)
	if err != nil || found != created {
		t.Fatalf("created token cannot authenticate: %+v, %v", found, err)
	}
	if _, err := store.FindByTokenHash(t.Context(), token); !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("raw token was stored instead of its hash: %v", err)
	}
	if err := svc.Revoke(t.Context(), token); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.FindByToken(t.Context(), token); !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("revoked token can still authenticate: %v", err)
	}
}
