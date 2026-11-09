package session

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"backend/internal/platform/database"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type errorRow struct{ err error }

func (r errorRow) Scan(...any) error { return r.err }

type testDB struct {
	database.DBTX
	err   error
	query string
	args  []any
}

func (db *testDB) QueryRow(_ context.Context, query string, args ...any) pgx.Row {
	db.query, db.args = query, args
	return errorRow{db.err}
}
func (db *testDB) Exec(_ context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	db.query, db.args = query, args
	return pgconn.CommandTag{}, db.err
}

func TestStoreErrorsAndExpiryFilter(t *testing.T) {
	ctx := context.Background()
	db := &testDB{err: pgx.ErrNoRows}
	store := NewStore(db)
	if _, err := store.FindByTokenHash(ctx, "hash"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("error = %v", err)
	}
	if !strings.Contains(db.query, `"expiresAt" > now()`) || db.args[0] != "hash" {
		t.Fatal("lookup must filter expired tokens")
	}
	failure := errors.New("database unavailable")
	db.err = failure
	if _, err := store.FindByTokenHash(ctx, "hash"); !errors.Is(err, failure) {
		t.Fatalf("lookup error = %v", err)
	}
	if _, err := store.Create(ctx, "user-id", "hash", time.Now()); !errors.Is(err, failure) {
		t.Fatalf("create error = %v", err)
	}
	if db.args[0] != "user-id" || db.args[1] != "hash" {
		t.Fatal("unexpected create arguments")
	}
	if err := store.DeleteByTokenHash(ctx, "hash"); !errors.Is(err, failure) {
		t.Fatalf("delete error = %v", err)
	}
	db.err = nil
	if err := store.DeleteByTokenHash(ctx, "hash"); err != nil {
		t.Fatal(err)
	}
}
