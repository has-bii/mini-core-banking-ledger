//go:build integration

package session_test

import (
	"backend/internal/modules/session"
	"backend/internal/testutil"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestStore_SessionLifecycle(t *testing.T) {
	t.Parallel()
	client := testutil.Redis(t)
	store := session.NewStore(client)
	userID := uuid.NewString()
	hash := strings.Repeat("a", 64)
	key := "session:" + hash
	expires := time.Now().UTC().Add(time.Hour)
	created, err := store.Create(t.Context(), userID, hash, expires)
	if err != nil {
		t.Fatal(err)
	}
	id, err := uuid.Parse(created.ID)
	if err != nil || id.Version() != 7 || created.UserID != userID || !created.ExpiresAt.Equal(expires) {
		t.Fatalf("Create() = %+v, want UUIDv7 session for user %s expiring %s", created, userID, expires)
	}
	ttl, err := client.PTTL(t.Context(), key).Result()
	if err != nil || ttl <= 0 || ttl > time.Until(expires)+time.Second {
		t.Fatalf("session TTL = %v, %v; want positive TTL bounded by expiration", ttl, err)
	}
	found, err := store.FindByTokenHash(t.Context(), hash)
	if err != nil || found != created {
		t.Fatalf("FindByTokenHash() = %+v, %v; want %+v", found, err, created)
	}
	if duplicate, err := store.Create(t.Context(), userID, hash, expires); err == nil || duplicate != (session.Session{}) {
		t.Fatalf("duplicate Create() = %+v, %v; want error", duplicate, err)
	}
	if found, err := store.FindByTokenHash(t.Context(), hash); err != nil || found != created {
		t.Fatalf("duplicate creation changed original session: %+v, %v", found, err)
	}
	for range 2 {
		if err := store.DeleteByTokenHash(t.Context(), hash); err != nil {
			t.Fatalf("deletion must be idempotent: %v", err)
		}
	}
	if _, err := store.FindByTokenHash(t.Context(), hash); !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("deleted session error = %v, want ErrNotFound", err)
	}

	// Force Redis expiry without sleeping or relying on scheduling.
	if _, err := store.Create(t.Context(), userID, hash, expires); err != nil {
		t.Fatal(err)
	}
	if err := client.ExpireAt(t.Context(), key, time.Now().Add(-time.Hour)).Err(); err != nil {
		t.Fatal(err)
	}
	if _, err := store.FindByTokenHash(t.Context(), hash); !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("Redis-expired session error = %v, want ErrNotFound", err)
	}
}

func TestStore_FindByTokenHash(t *testing.T) {
	t.Parallel()
	encode := func(value session.Session) string {
		t.Helper()
		data, err := json.Marshal(value)
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	valid := session.Session{ID: uuid.NewString(), UserID: uuid.NewString(), ExpiresAt: time.Now().UTC().Add(time.Hour)}
	expired := valid
	expired.ExpiresAt = time.Now().UTC().Add(-time.Hour)
	for _, tt := range []struct {
		name, payload string
		missing       bool
		want          session.Session
		wantError     bool
		wantNotFound  bool
	}{
		{name: "valid", payload: encode(valid), want: valid},
		{name: "missing", missing: true, wantError: true, wantNotFound: true},
		{name: "expired payload without TTL", payload: encode(expired), wantError: true, wantNotFound: true},
		{name: "invalid JSON", payload: "{", wantError: true},
		{name: "null", payload: "null", wantError: true},
		{name: "missing fields", payload: "{}", wantError: true},
		{name: "missing id", payload: encode(session.Session{UserID: valid.UserID, ExpiresAt: valid.ExpiresAt}), wantError: true},
		{name: "missing user", payload: encode(session.Session{ID: valid.ID, ExpiresAt: valid.ExpiresAt}), wantError: true},
		{name: "missing expiration", payload: encode(session.Session{ID: valid.ID, UserID: valid.UserID}), wantError: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			client := testutil.Redis(t)
			hash := strings.Repeat("b", 64)
			if !tt.missing {
				if err := client.Set(t.Context(), "session:"+hash, tt.payload, 0).Err(); err != nil {
					t.Fatal(err)
				}
			}
			got, err := session.NewStore(client).FindByTokenHash(t.Context(), hash)
			if got != tt.want || (err != nil) != tt.wantError || errors.Is(err, session.ErrNotFound) != tt.wantNotFound {
				t.Fatalf("FindByTokenHash() = %+v, %v; want %+v, error=%v, notFound=%v", got, err, tt.want, tt.wantError, tt.wantNotFound)
			}
		})
	}
}
