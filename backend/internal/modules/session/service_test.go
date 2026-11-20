package session_test

import (
	"backend/internal/modules/session"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"
)

type sessionStoreStub struct {
	create func(context.Context, string, string, time.Time) (session.Session, error)
	get    func(context.Context, string) (session.Session, error)
	delete func(context.Context, string) error
}

func (s sessionStoreStub) Create(ctx context.Context, userID, tokenHash string, expiresAt time.Time) (session.Session, error) {
	return s.create(ctx, userID, tokenHash, expiresAt)
}

func (s sessionStoreStub) GetByTokenHash(ctx context.Context, tokenHash string) (session.Session, error) {
	return s.get(ctx, tokenHash)
}

func (s sessionStoreStub) DeleteByTokenHash(ctx context.Context, tokenHash string) error {
	return s.delete(ctx, tokenHash)
}

const (
	knownToken = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
	knownHash  = "0f007385b6f9d4b7eeb2748605afe1a984a0a3bfa3f014d09e2a784ce9e5cd1a"
)

func TestService_Create(t *testing.T) {
	t.Parallel()
	const userID = "019a3de4-8750-7000-8000-000000000001"
	const lifetime = 7 * time.Minute
	for _, tt := range []struct {
		name, userID string
		cause        error
	}{
		{name: "success", userID: userID},
		{name: "invalid user ID", userID: "not-a-uuid", cause: session.ErrInvalidUserID},
		{name: "store failure", userID: userID, cause: errors.New("session store unavailable")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			before := time.Now()
			var storedHash string
			var created session.Session
			store := sessionStoreStub{create: func(gotCtx context.Context, gotUserID, tokenHash string, expiresAt time.Time) (session.Session, error) {
				if gotCtx != ctx || gotUserID != tt.userID {
					t.Error("Create() did not receive the caller context and user ID")
				}
				if expiresAt.Before(before.Add(lifetime)) || expiresAt.After(time.Now().Add(lifetime)) {
					t.Errorf("Create() received expiration %v; want now plus %v", expiresAt, lifetime)
				}
				storedHash = tokenHash
				created = session.Session{ID: "session-id", UserID: gotUserID, ExpiresAt: expiresAt}
				return created, tt.cause
			}}

			got, token, err := session.NewService(store, lifetime).Create(ctx, tt.userID)
			if !errors.Is(err, tt.cause) {
				t.Fatalf("Create() error = %v; want %v", err, tt.cause)
			}
			if tt.cause != nil {
				if got != (session.Session{}) || token != "" {
					t.Error("failed Create() returned a session or token")
				}
				return
			}
			if got != created {
				t.Error("Create() did not return the stored session")
			}
			if decoded, err := base64.RawURLEncoding.Strict().DecodeString(token); err != nil || len(decoded) != 32 || len(token) != 43 {
				t.Fatal("Create() token is not a canonical 32-byte base64url value")
			}
			digest := sha256.Sum256([]byte(token))
			if storedHash != hex.EncodeToString(digest[:]) {
				t.Error("Create() did not pass the SHA-256 digest of the raw token to the store")
			}
		})
	}
}

func TestService_GetByToken(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name  string
		want  session.Session
		cause error
	}{
		{name: "preserves stored expiry", want: session.Session{ID: "session-id", UserID: "user-id", ExpiresAt: time.Now().Add(2 * time.Hour)}},
		{name: "not found", cause: session.ErrNotFound},
		{name: "store failure", cause: errors.New("session store unavailable")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			store := sessionStoreStub{get: func(gotCtx context.Context, tokenHash string) (session.Session, error) {
				if gotCtx != ctx || tokenHash != knownHash {
					t.Error("GetByTokenHash() did not receive the caller context and expected token hash")
				}
				return tt.want, tt.cause
			}}
			got, err := session.NewService(store, time.Hour).GetByToken(ctx, knownToken)
			if got != tt.want || !errors.Is(err, tt.cause) {
				t.Errorf("GetByToken() = %+v, %v; want %+v, %v", got, err, tt.want, tt.cause)
			}
		})
	}
}

func TestService_Revoke(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name  string
		cause error
	}{
		{name: "success"},
		{name: "store failure", cause: errors.New("session store unavailable")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()
			called := false
			store := sessionStoreStub{delete: func(gotCtx context.Context, tokenHash string) error {
				called = true
				if gotCtx != ctx || tokenHash != knownHash {
					t.Error("DeleteByTokenHash() did not receive the caller context and expected token hash")
				}
				return tt.cause
			}}
			if err := session.NewService(store, time.Hour).Revoke(ctx, knownToken); !errors.Is(err, tt.cause) || !called {
				t.Errorf("Revoke() error = %v, called DeleteByTokenHash() = %t; want %v and a deletion attempt", err, called, tt.cause)
			}
		})
	}
}

func TestService_RejectsMalformedTokens(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, token string
	}{
		{"empty", ""},
		{"short", strings.Repeat("A", 42)},
		{"padded", strings.Repeat("A", 43) + "="},
		{"invalid alphabet", strings.Repeat("!", 43)},
		{"noncanonical trailing bits", strings.Repeat("A", 42) + "B"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// A nil store proves malformed input is rejected before persistence.
			svc := session.NewService(nil, time.Hour)
			got, err := svc.GetByToken(t.Context(), tt.token)
			if !errors.Is(err, session.ErrNotFound) || got != (session.Session{}) {
				t.Fatalf("GetByToken() = %+v, %v; want not found", got, err)
			}
			if err := svc.Revoke(t.Context(), tt.token); !errors.Is(err, session.ErrNotFound) {
				t.Fatalf("Revoke() = %v, want not found", err)
			}
		})
	}
}
