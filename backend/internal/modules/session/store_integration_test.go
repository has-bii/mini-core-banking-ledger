//go:build integration

package session_test

import (
	"backend/internal/modules/session"
	"backend/internal/modules/user"
	"backend/internal/testutil"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestStore_SessionLifecycle(t *testing.T) {
	t.Parallel()
	pool := testutil.Postgres(t)
	u, err := user.NewStore(pool).Create(t.Context(), "Alice", "alice@example.com")
	if err != nil {
		t.Fatal(err)
	}
	store := session.NewStore(pool)
	hash := strings.Repeat("a", 64)
	expires := time.Now().UTC().Add(time.Hour).Truncate(time.Microsecond)
	created, err := store.Create(t.Context(), u.ID, hash, expires)
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == "" || created.UserID != u.ID || !created.ExpiresAt.Equal(expires) {
		t.Fatalf("Create() = %+v, want session for user %s expiring %s", created, u.ID, expires)
	}
	found, err := store.FindByTokenHash(t.Context(), hash)
	if err != nil || found != created {
		t.Fatalf("FindByTokenHash() = %+v, %v; want %+v", found, err, created)
	}
	if err := store.DeleteByTokenHash(t.Context(), hash); err != nil {
		t.Fatal(err)
	}
	if _, err := store.FindByTokenHash(t.Context(), hash); !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("deleted session error = %v, want ErrNotFound", err)
	}
	if err := store.DeleteByTokenHash(t.Context(), hash); err != nil {
		t.Fatalf("repeated deletion must succeed: %v", err)
	}
	if _, err := store.Create(t.Context(), u.ID, hash, time.Now().Add(-time.Hour)); err != nil {
		t.Fatal(err)
	}
	if _, err := store.FindByTokenHash(t.Context(), hash); !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("expired session error = %v, want ErrNotFound", err)
	}
	if _, err := store.FindByTokenHash(t.Context(), strings.Repeat("b", 64)); !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("unknown session error = %v, want ErrNotFound", err)
	}
}
