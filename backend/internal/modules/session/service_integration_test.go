//go:build integration

package session_test

import (
	"backend/internal/modules/session"
	"backend/internal/modules/user"
	"backend/internal/testutil"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestService_StoreErrors(t *testing.T) {
	t.Parallel()
	pool := testutil.Postgres(t)
	svc := session.NewService(session.NewStore(pool))
	pool.Close()
	failure := pool.Ping(t.Context())
	if failure == nil {
		t.Fatal("closed pool must fail")
	}
	got, token, err := svc.Create(t.Context(), "user-id")
	if !errors.Is(err, failure) || got != (session.Session{}) || token != "" {
		t.Fatalf("Create() = %+v, %q, %v; want no session/token and store error", got, token, err)
	}
	got, err = svc.FindByToken(t.Context(), strings.Repeat("A", 43))
	if !errors.Is(err, failure) || got != (session.Session{}) {
		t.Fatalf("FindByToken() = %+v, %v; want no session and store error", got, err)
	}
	if err := svc.Revoke(t.Context(), strings.Repeat("A", 43)); !errors.Is(err, failure) {
		t.Fatalf("Revoke() = %v, want store error", err)
	}
}

func TestService_TokenPersistence(t *testing.T) {
	t.Parallel()
	pool := testutil.Postgres(t)
	u, err := user.NewStore(pool).Create(t.Context(), "Alice", "alice@example.com", "stored-password-hash")
	if err != nil {
		t.Fatal(err)
	}
	store := session.NewStore(pool)
	svc := session.NewService(store)
	before := time.Now().UTC().Add(24 * time.Hour).Truncate(time.Microsecond)
	created, token, err := svc.Create(t.Context(), u.ID)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(token)
	if err != nil || len(decoded) != 32 || len(token) != 43 {
		t.Fatalf("token is not a canonical 32-byte base64url value: %q (%v)", token, err)
	}
	if created.ID == "" || created.UserID != u.ID || created.ExpiresAt.Before(before) || created.ExpiresAt.After(time.Now().UTC().Add(24*time.Hour)) {
		t.Fatalf("Create() = %+v, want a stored session with a 24-hour lifetime", created)
	}
	digest := sha256.Sum256([]byte(token))
	stored, err := store.FindByTokenHash(t.Context(), hex.EncodeToString(digest[:]))
	if err != nil || stored != created {
		t.Fatalf("session not stored under SHA-256 hash: %+v, %v", stored, err)
	}
	if _, err := store.FindByTokenHash(t.Context(), token); !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("raw token must not be persisted: %v", err)
	}
	if found, err := svc.FindByToken(t.Context(), token); err != nil || found != created {
		t.Fatalf("FindByToken() = %+v, %v; want %+v", found, err, created)
	}
	if err := svc.Revoke(t.Context(), token); err != nil {
		t.Fatal(err)
	}
	if _, err := store.FindByTokenHash(t.Context(), hex.EncodeToString(digest[:])); !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("Revoke() did not delete the hashed token: %v", err)
	}

	// A known token independently checks lookup and revocation hashing.
	knownToken := strings.Repeat("A", 43)
	const knownHash = "0f007385b6f9d4b7eeb2748605afe1a984a0a3bfa3f014d09e2a784ce9e5cd1a"
	known, err := store.Create(t.Context(), u.ID, knownHash, time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if found, err := svc.FindByToken(t.Context(), knownToken); err != nil || found != known {
		t.Fatalf("known token lookup = %+v, %v", found, err)
	}
	if err := svc.Revoke(t.Context(), knownToken); err != nil {
		t.Fatal(err)
	}
	if _, err := store.FindByTokenHash(t.Context(), knownHash); !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("known token revocation did not delete hash: %v", err)
	}
}
