package session

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"strings"
	"testing"
	"time"
)

type fakeStore struct {
	create func(context.Context, string, string, time.Time) (Session, error)
	find   func(context.Context, string) (Session, error)
	delete func(context.Context, string) error
}

func (f fakeStore) Create(ctx context.Context, userID, hash string, expires time.Time) (Session, error) {
	return f.create(ctx, userID, hash, expires)
}

func (f fakeStore) FindByTokenHash(ctx context.Context, hash string) (Session, error) {
	return f.find(ctx, hash)
}

func (f fakeStore) DeleteByTokenHash(ctx context.Context, hash string) error {
	return f.delete(ctx, hash)
}

func TestCreate(t *testing.T) {
	failure := errors.New("store unavailable")
	for _, storeErr := range []error{nil, failure} {
		name := "success"
		if storeErr != nil {
			name = "store failure"
		}
		t.Run(name, func(t *testing.T) {
			var storedHash string
			var want Session
			before := time.Now().Add(Lifetime)
			service := NewService(fakeStore{create: func(_ context.Context, userID, hash string, expires time.Time) (Session, error) {
				if userID != "user-1" {
					t.Errorf("user ID = %q", userID)
				}
				if expires.Before(before) || expires.After(time.Now().Add(Lifetime)) {
					t.Errorf("expiry = %v, want now + %v", expires, Lifetime)
				}
				storedHash = hash
				want = Session{ID: "session-1", UserID: userID, ExpiresAt: expires}
				return want, storeErr
			}})
			got, token, err := service.Create(context.Background(), "user-1")
			if !errors.Is(err, storeErr) {
				t.Fatalf("Create() error = %v, want %v", err, storeErr)
			}
			if storeErr != nil {
				if got != (Session{}) || token != "" {
					t.Error("failed creation returned session or token")
				}
				return
			}
			if got != want {
				t.Errorf("session = %v, want %v", got, want)
			}
			decoded, err := base64.RawURLEncoding.Strict().DecodeString(token)
			if err != nil || len(decoded) != 32 {
				t.Fatalf("token must encode 32 bytes: %q, %v", token, err)
			}
			digest := sha256.Sum256([]byte(token))
			if storedHash != hex.EncodeToString(digest[:]) || storedHash == token {
				t.Error("store must receive the SHA-256 hash, not the raw token")
			}
		})
	}
}

func TestTokenOperations(t *testing.T) {
	valid := base64.RawURLEncoding.EncodeToString(make([]byte, 32))
	failure := errors.New("store unavailable")
	for _, tt := range []struct {
		name     string
		token    string
		storeErr error
		wantErr  error
	}{
		{"valid", valid, nil, nil},
		{"missing", valid, ErrNotFound, ErrNotFound},
		{"store failure", valid, failure, failure},
		{"empty", "", nil, ErrNotFound},
		{"short", valid[:42], nil, ErrNotFound},
		{"long", valid + "A", nil, ErrNotFound},
		{"invalid alphabet", strings.Repeat("!", 43), nil, ErrNotFound},
		{"noncanonical padding bits", strings.Repeat("A", 42) + "B", nil, ErrNotFound},
	} {
		t.Run(tt.name, func(t *testing.T) {
			findCalls, deleteCalls := 0, 0
			want := Session{ID: "session-1", UserID: "user-1"}
			digest := sha256.Sum256([]byte(tt.token))
			wantHash := hex.EncodeToString(digest[:])
			service := NewService(fakeStore{
				find: func(_ context.Context, hash string) (Session, error) {
					findCalls++
					if hash != wantHash {
						t.Errorf("lookup hash = %q, want %q", hash, wantHash)
					}
					if tt.storeErr != nil {
						return Session{}, tt.storeErr
					}
					return want, nil
				},
				delete: func(_ context.Context, hash string) error {
					deleteCalls++
					if hash != wantHash {
						t.Errorf("revocation hash = %q, want %q", hash, wantHash)
					}
					return tt.storeErr
				},
			})
			got, err := service.FindByToken(context.Background(), tt.token)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("FindByToken() error = %v, want %v", err, tt.wantErr)
			}
			if tt.wantErr == nil && got != want || tt.wantErr != nil && got != (Session{}) {
				t.Errorf("unexpected session: %v", got)
			}
			if err := service.Revoke(context.Background(), tt.token); !errors.Is(err, tt.wantErr) {
				t.Errorf("Revoke() error = %v, want %v", err, tt.wantErr)
			}
			wantCalls := 0
			if tt.token == valid {
				wantCalls = 1
			}
			if findCalls != wantCalls || deleteCalls != wantCalls {
				t.Errorf("store calls = (%d, %d), want (%d, %d)", findCalls, deleteCalls, wantCalls, wantCalls)
			}
		})
	}
}
