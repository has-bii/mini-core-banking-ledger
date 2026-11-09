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
	find   func(context.Context, string) (session.Session, error)
	delete func(context.Context, string) error
}

func (s sessionStoreStub) Create(ctx context.Context, userID, hash string, expires time.Time) (session.Session, error) {
	return s.create(ctx, userID, hash, expires)
}
func (s sessionStoreStub) FindByTokenHash(ctx context.Context, hash string) (session.Session, error) {
	return s.find(ctx, hash)
}
func (s sessionStoreStub) DeleteByTokenHash(ctx context.Context, hash string) error {
	return s.delete(ctx, hash)
}

func TestService_Create(t *testing.T) {
	t.Parallel()
	var storedHash string
	before := time.Now().UTC().Add(24 * time.Hour)
	svc := session.NewService(sessionStoreStub{
		create: func(_ context.Context, userID, hash string, expires time.Time) (session.Session, error) {
			storedHash = hash
			return session.Session{ID: "session-id", UserID: userID, ExpiresAt: expires}, nil
		},
	})
	got, token, err := svc.Create(t.Context(), "user-id")
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := base64.RawURLEncoding.Strict().DecodeString(token)
	if err != nil || len(decoded) != 32 || len(token) != 43 {
		t.Fatalf("token is not a canonical 32-byte base64url value: %q (%v)", token, err)
	}
	digest := sha256.Sum256([]byte(token))
	if storedHash == token || storedHash != hex.EncodeToString(digest[:]) {
		t.Error("persistence did not receive the SHA-256 hash of the raw token")
	}
	if got.ID != "session-id" || got.UserID != "user-id" || got.ExpiresAt.Before(before) || got.ExpiresAt.After(time.Now().UTC().Add(24*time.Hour)) {
		t.Errorf("Create() = %+v, want stored session with a 24-hour lifetime", got)
	}
}

func TestService_CreateStoreError(t *testing.T) {
	t.Parallel()
	failure := errors.New("store unavailable")
	svc := session.NewService(sessionStoreStub{
		create: func(context.Context, string, string, time.Time) (session.Session, error) {
			return session.Session{}, failure
		},
	})
	got, token, err := svc.Create(t.Context(), "user-id")
	if !errors.Is(err, failure) || got != (session.Session{}) || token != "" {
		t.Fatalf("Create() = %+v, %q, %v; want no session/token and store error", got, token, err)
	}
}

func TestService_FindByToken(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name  string
		token string
		valid bool
	}{
		{"valid", strings.Repeat("A", 43), true},
		{"empty", "", false},
		{"short", strings.Repeat("A", 42), false},
		{"padded", strings.Repeat("A", 43) + "=", false},
		{"invalid alphabet", strings.Repeat("!", 43), false},
		{"noncanonical trailing bits", strings.Repeat("A", 42) + "B", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := session.NewService(sessionStoreStub{
				find: func(_ context.Context, hash string) (session.Session, error) {
					if !tt.valid {
						t.Fatal("malformed token reached persistence")
					}
					if hash != "0f007385b6f9d4b7eeb2748605afe1a984a0a3bfa3f014d09e2a784ce9e5cd1a" {
						t.Errorf("unexpected token hash: %q", hash)
					}
					return session.Session{ID: "session-id"}, nil
				},
			})
			got, err := svc.FindByToken(t.Context(), tt.token)
			if tt.valid {
				if err != nil || got.ID != "session-id" {
					t.Fatalf("FindByToken() = %+v, %v", got, err)
				}
			} else if !errors.Is(err, session.ErrNotFound) || got != (session.Session{}) {
				t.Fatalf("FindByToken() = %+v, %v; want not found", got, err)
			}
		})
	}
}

func TestService_Revoke(t *testing.T) {
	t.Parallel()
	failure := errors.New("delete failed")
	for _, tt := range []struct {
		name     string
		token    string
		storeErr error
		wantErr  error
	}{
		{"valid", strings.Repeat("A", 43), nil, nil},
		{"malformed", "bad-token", nil, session.ErrNotFound},
		{"store error", strings.Repeat("A", 43), failure, failure},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := session.NewService(sessionStoreStub{
				delete: func(_ context.Context, hash string) error {
					if tt.token == "bad-token" {
						t.Fatal("malformed token reached persistence")
					}
					if hash != "0f007385b6f9d4b7eeb2748605afe1a984a0a3bfa3f014d09e2a784ce9e5cd1a" {
						t.Errorf("unexpected token hash: %q", hash)
					}
					return tt.storeErr
				},
			})
			if err := svc.Revoke(t.Context(), tt.token); !errors.Is(err, tt.wantErr) {
				t.Fatalf("Revoke() = %v, want %v", err, tt.wantErr)
			}
		})
	}
}
