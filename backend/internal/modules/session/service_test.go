package session

import (
	"context"
	"errors"
	"testing"
	"time"
)

type memoryStore struct {
	hash    string
	session Session
	err     error
}

func (s *memoryStore) Create(_ context.Context, userID, hash string, expires time.Time) (Session, error) {
	s.hash = hash
	s.session = Session{ID: "session-id", UserID: userID, ExpiresAt: expires}
	return s.session, s.err
}
func (s *memoryStore) FindByTokenHash(_ context.Context, hash string) (Session, error) {
	if hash != s.hash || !s.session.ExpiresAt.After(time.Now()) {
		return Session{}, ErrNotFound
	}
	return s.session, s.err
}
func (s *memoryStore) DeleteByTokenHash(_ context.Context, hash string) error {
	if hash == s.hash {
		s.hash = ""
	}
	return s.err
}

func TestSessionLifecycle(t *testing.T) {
	ctx := context.Background()
	store := &memoryStore{}
	svc := NewService(store)
	before := time.Now()
	created, token, err := svc.Create(ctx, "user-id")
	if err != nil {
		t.Fatal(err)
	}
	if !validToken(token) || store.hash == token || store.hash != hashToken(token) {
		t.Fatal("expected random token with only its hash persisted")
	}
	if created.UserID != "user-id" || created.ExpiresAt.Before(before.Add(Lifetime)) || created.ExpiresAt.After(time.Now().Add(Lifetime)) {
		t.Fatalf("unexpected session: %+v", created)
	}
	if got, err := svc.FindByToken(ctx, token); err != nil || got != created {
		t.Fatalf("lookup = %+v, %v", got, err)
	}
	if err := svc.Revoke(ctx, token); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.FindByToken(ctx, token); !errors.Is(err, ErrNotFound) {
		t.Fatalf("revoked session error = %v", err)
	}
	_, nextToken, err := svc.Create(ctx, "user-id")
	if err != nil || nextToken == token {
		t.Fatal("tokens must be unique")
	}
	store.session.ExpiresAt = time.Now().Add(-time.Second)
	if _, err := svc.FindByToken(ctx, nextToken); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expired session error = %v", err)
	}
	for _, invalid := range []string{"", "bad", token + "="} {
		if _, err := svc.FindByToken(ctx, invalid); !errors.Is(err, ErrNotFound) {
			t.Errorf("accepted %q", invalid)
		}
		if err := svc.Revoke(ctx, invalid); !errors.Is(err, ErrNotFound) {
			t.Errorf("revoked %q", invalid)
		}
	}
}

func TestCreateFailureDoesNotReturnToken(t *testing.T) {
	failure := errors.New("database unavailable")
	svc := NewService(&memoryStore{err: failure})
	_, token, err := svc.Create(context.Background(), "user-id")
	if !errors.Is(err, failure) || token != "" {
		t.Fatalf("token = %q, error = %v", token, err)
	}
}
