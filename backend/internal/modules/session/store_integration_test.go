//go:build integration

package session_test

import (
	"backend/internal/modules/session"
	"backend/internal/testutil"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
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
		t.Fatal("Create() did not return a UUIDv7 session with the supplied owner and expiration")
	}
	data, err := client.Get(t.Context(), key).Bytes()
	if err != nil {
		t.Fatal(err)
	}
	var persisted session.Session
	if err := json.Unmarshal(data, &persisted); err != nil {
		t.Fatal(err)
	}
	want := session.Session{ID: created.ID, UserID: userID, ExpiresAt: expires}
	if persisted != want {
		t.Error("stored JSON does not match the expected session fields")
	}
	ttl, err := client.PTTL(t.Context(), key).Result()
	if err != nil || ttl <= 0 || ttl > time.Until(expires)+time.Second {
		t.Fatalf("session TTL = %v, %v; want positive TTL bounded by expiration", ttl, err)
	}
	found, err := store.GetByTokenHash(t.Context(), hash)
	if err != nil || found != want {
		t.Fatalf("GetByTokenHash() error = %v; want the persisted session", err)
	}
	if duplicate, err := store.Create(t.Context(), uuid.NewString(), hash, expires.Add(time.Hour)); !errors.Is(err, session.ErrTokenHashTaken) || duplicate != (session.Session{}) {
		t.Fatalf("duplicate Create() error = %v; want zero session and ErrTokenHashTaken", err)
	}
	if found, err := store.GetByTokenHash(t.Context(), hash); err != nil || found != want {
		t.Fatalf("duplicate creation changed the original session: %v", err)
	}
	if ttl, err := client.PTTL(t.Context(), key).Result(); err != nil || ttl <= 0 || ttl > time.Until(expires)+time.Second {
		t.Fatalf("collision changed original TTL: %v, %v", ttl, err)
	}
	for range 2 {
		if err := store.DeleteByTokenHash(t.Context(), hash); err != nil {
			t.Fatalf("deletion must be idempotent: %v", err)
		}
	}
	if got, err := store.GetByTokenHash(t.Context(), hash); got != (session.Session{}) || !errors.Is(err, session.ErrNotFound) || errors.Is(err, redis.Nil) {
		t.Fatalf("deleted session error = %v, want zero session and domain ErrNotFound", err)
	}

	// Force Redis expiry without sleeping or relying on scheduling.
	if _, err := store.Create(t.Context(), userID, hash, expires); err != nil {
		t.Fatal(err)
	}
	if err := client.ExpireAt(t.Context(), key, time.Now().Add(-time.Hour)).Err(); err != nil {
		t.Fatal(err)
	}
	if got, err := store.GetByTokenHash(t.Context(), hash); got != (session.Session{}) || !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("Redis-expired session error = %v, want zero session and ErrNotFound", err)
	}
}

func TestStore_Create_InvalidArguments(t *testing.T) {
	t.Parallel()
	client := testutil.Redis(t)
	store := session.NewStore(client)
	for _, tt := range []struct {
		name, userID string
		expiresAt    time.Time
		want         error
	}{
		{"empty user ID", "", time.Now().Add(time.Hour), session.ErrInvalidUserID},
		{"whitespace user ID", " ", time.Now().Add(time.Hour), session.ErrInvalidUserID},
		{"malformed user ID", "not-a-uuid", time.Now().Add(time.Hour), session.ErrInvalidUserID},
		{"invalid UUID digit", "019a3de4-8750-7000-8000-00000000000z", time.Now().Add(time.Hour), session.ErrInvalidUserID},
		{"zero expiration", uuid.NewString(), time.Time{}, session.ErrInvalidExpiration},
		{"past expiration", uuid.NewString(), time.Now().Add(-time.Hour), session.ErrInvalidExpiration},
	} {
		t.Run(tt.name, func(t *testing.T) {
			hash := strings.Repeat("b", 64)
			got, err := store.Create(t.Context(), tt.userID, hash, tt.expiresAt)
			if got != (session.Session{}) || !errors.Is(err, tt.want) {
				t.Fatalf("Create() error = %v, want zero session and %v", err, tt.want)
			}
			if count, err := client.Exists(t.Context(), "session:"+hash).Result(); err != nil || count != 0 {
				t.Fatalf("invalid creation persisted a key: count=%d, error=%v", count, err)
			}
		})
	}
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Create(t.Context(), "not-a-uuid", strings.Repeat("c", 64), time.Now().Add(time.Hour)); !errors.Is(err, session.ErrInvalidUserID) {
		t.Fatalf("identifier validation must precede Redis access: %v", err)
	}
	if _, err := store.Create(t.Context(), uuid.NewString(), strings.Repeat("c", 64), time.Time{}); !errors.Is(err, session.ErrInvalidExpiration) {
		t.Fatalf("expiration validation must precede Redis access: %v", err)
	}
}

func TestStore_GetByTokenHash(t *testing.T) {
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
		name, payload, cause string
		missing              bool
		want                 session.Session
		wantError            bool
		wantNotFound         bool
	}{
		{name: "existing session format", payload: fmt.Sprintf(`{"id":%q,"user_id":%q,"expires_at":%q}`,
			valid.ID, valid.UserID, valid.ExpiresAt.Format(time.RFC3339Nano)), want: valid},
		{name: "missing", missing: true, wantError: true, wantNotFound: true},
		{name: "expired payload without TTL", payload: encode(expired), wantError: true, wantNotFound: true},
		{name: "invalid JSON", payload: `{"user_id":]}`, wantError: true, cause: "syntax"},
		{name: "null", payload: "null", wantError: true},
		{name: "missing fields", payload: "{}", wantError: true},
		{name: "missing id", payload: encode(session.Session{UserID: valid.UserID, ExpiresAt: valid.ExpiresAt}), wantError: true},
		{name: "missing user", payload: encode(session.Session{ID: valid.ID, ExpiresAt: valid.ExpiresAt}), wantError: true},
		{name: "missing expiration", payload: encode(session.Session{ID: valid.ID, UserID: valid.UserID}), wantError: true},
		{name: "wrong field type", payload: `{"user_id":123}`, wantError: true, cause: "type"},
		{name: "malformed expiration", payload: `{"expires_at":"not-a-time"}`, wantError: true, cause: "time"},
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
			got, err := session.NewStore(client).GetByTokenHash(t.Context(), hash)
			if got != tt.want || (err != nil) != tt.wantError || errors.Is(err, session.ErrNotFound) != tt.wantNotFound || errors.Is(err, redis.Nil) {
				t.Fatalf("GetByTokenHash() error = %v; want error=%v, notFound=%v and expected result", err, tt.wantError, tt.wantNotFound)
			}
			switch tt.cause {
			case "syntax":
				var cause *jsontext.SyntacticError
				if !errors.As(err, &cause) || !errors.Is(err, cause) {
					t.Fatalf("error = %v, want preserved JSON syntax cause", err)
				}
			case "type":
				var cause *json.SemanticError
				if !errors.As(err, &cause) || !errors.Is(err, cause) {
					t.Fatalf("error = %v, want preserved JSON type cause", err)
				}
			case "time":
				var cause *time.ParseError
				if !errors.As(err, &cause) || !errors.Is(err, cause) {
					t.Fatalf("error = %v, want preserved expiration decoding cause", err)
				}
			}
		})
	}
}

func TestStore_UnexpectedFailures(t *testing.T) {
	t.Parallel()
	client := testutil.Redis(t)
	store := session.NewStore(client)
	if err := client.Close(); err != nil {
		t.Fatal(err)
	}
	hash := strings.Repeat("c", 64)
	got, err := store.Create(t.Context(), uuid.NewString(), hash, time.Now().Add(time.Hour))
	if got != (session.Session{}) || !errors.Is(err, redis.ErrClosed) {
		t.Fatalf("Create() error = %v, want zero session and preserved Redis failure", err)
	}
	got, err = store.GetByTokenHash(t.Context(), hash)
	if got != (session.Session{}) || !errors.Is(err, redis.ErrClosed) {
		t.Fatalf("GetByTokenHash() error = %v, want zero session and preserved Redis failure", err)
	}
	if err := store.DeleteByTokenHash(t.Context(), hash); !errors.Is(err, redis.ErrClosed) {
		t.Fatalf("DeleteByTokenHash() error = %v, want preserved Redis failure", err)
	}
}

func TestStore_Create_ConcurrentCollision(t *testing.T) {
	t.Parallel()
	client := testutil.Redis(t)
	store := session.NewStore(client)
	hash := strings.Repeat("d", 64)
	expires := time.Now().UTC().Add(time.Hour)
	type result struct {
		session session.Session
		userID  string
		err     error
	}
	start := make(chan struct{})
	results := make(chan result, 2)
	for range 2 {
		userID := uuid.NewString()
		go func() {
			<-start
			created, err := store.Create(t.Context(), userID, hash, expires)
			results <- result{created, userID, err}
		}()
	}
	close(start)
	var winner session.Session
	var successes, conflicts int
	for range 2 {
		r := <-results
		switch {
		case r.err == nil && r.session.ID != "" && r.session.UserID == r.userID && r.session.ExpiresAt.Equal(expires):
			winner = r.session
			successes++
		case errors.Is(r.err, session.ErrTokenHashTaken) && r.session == (session.Session{}):
			conflicts++
		default:
			t.Errorf("unexpected concurrent creation error: %v", r.err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Fatalf("concurrent creation = %d successes, %d conflicts; want one each", successes, conflicts)
	}
	if got, err := store.GetByTokenHash(t.Context(), hash); err != nil || got != winner {
		t.Fatalf("stored session differs from the winning creation: %v", err)
	}
	if ttl, err := client.PTTL(t.Context(), "session:"+hash).Result(); err != nil || ttl <= 0 || ttl > time.Until(expires)+time.Second {
		t.Fatalf("concurrent creation TTL = %v, %v; want positive TTL bounded by expiration", ttl, err)
	}
}
