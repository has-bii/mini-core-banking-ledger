package auth

import (
	"backend/internal/modules/user"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
)

type registrationDB struct {
	pgx.Tx
	lookupErr, createErr, accountErr, beginErr, commitErr error
	hash                                                  string
	queries                                               int
	begun, committed, rolledBack                          bool
	userID                                                string
}

type registrationRow func(...any) error

func (r registrationRow) Scan(dest ...any) error { return r(dest...) }

func (db *registrationDB) Begin(context.Context) (pgx.Tx, error) {
	db.begun = true
	return db, db.beginErr
}
func (db *registrationDB) Commit(context.Context) error {
	db.committed = true
	return db.commitErr
}
func (db *registrationDB) Rollback(context.Context) error {
	db.rolledBack = true
	return nil
}
func (db *registrationDB) QueryRow(_ context.Context, query string, args ...any) pgx.Row {
	db.queries++
	return registrationRow(func(dest ...any) error {
		switch {
		case strings.HasPrefix(query, "SELECT"):
			return db.lookupErr
		case strings.Contains(query, "INSERT INTO users"):
			if db.createErr != nil {
				return db.createErr
			}
			*dest[0].(*string) = "user-id"
			*dest[1].(*string) = args[0].(string)
			*dest[2].(*string) = args[1].(string)
		}
		return nil
	})
}

func (db *registrationDB) Exec(_ context.Context, _ string, args ...any) (pgconn.CommandTag, error) {
	db.queries++
	db.userID = args[0].(string)
	db.hash = args[1].(string)
	return pgconn.CommandTag{}, db.accountErr
}

func TestRegister(t *testing.T) {
	dbErr := errors.New("database unavailable")
	for _, tc := range []struct {
		name                                                           string
		lookupErr, createErr, accountErr, beginErr, commitErr, wantErr error
		wantQueries                                                    int
		wantBegin, wantCommit, wantRollback                            bool
	}{
		{name: "success", lookupErr: pgx.ErrNoRows, wantQueries: 3, wantBegin: true, wantCommit: true, wantRollback: true},
		{name: "existing email", wantErr: user.ErrEmailTaken, wantQueries: 1},
		{name: "lookup failure", lookupErr: dbErr, wantErr: dbErr, wantQueries: 1},
		{name: "begin failure", lookupErr: pgx.ErrNoRows, beginErr: dbErr, wantErr: dbErr, wantQueries: 1, wantBegin: true},
		{name: "create failure", lookupErr: pgx.ErrNoRows, createErr: dbErr, wantErr: dbErr, wantQueries: 2, wantBegin: true, wantRollback: true},
		{name: "concurrent duplicate", lookupErr: pgx.ErrNoRows, createErr: &pgconn.PgError{Code: "23505", ConstraintName: "users_email_key"}, wantErr: user.ErrEmailTaken, wantQueries: 2, wantBegin: true, wantRollback: true},
		{name: "account failure", lookupErr: pgx.ErrNoRows, accountErr: dbErr, wantErr: dbErr, wantQueries: 3, wantBegin: true, wantRollback: true},
		{name: "commit failure", lookupErr: pgx.ErrNoRows, commitErr: dbErr, wantErr: dbErr, wantQueries: 3, wantBegin: true, wantCommit: true, wantRollback: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := &registrationDB{lookupErr: tc.lookupErr, createErr: tc.createErr, accountErr: tc.accountErr, beginErr: tc.beginErr, commitErr: tc.commitErr}
			s := NewService(user.NewStore(db))
			got, err := s.Register(t.Context(), RegisterPayload{"Alice", "alice@example.com", " password123 "})
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("error = %v, want %v", err, tc.wantErr)
			}
			if db.queries != tc.wantQueries || db.begun != tc.wantBegin || db.committed != tc.wantCommit || db.rolledBack != tc.wantRollback {
				t.Fatalf("queries = %d, begun = %v, committed = %v, rollback cleanup = %v", db.queries, db.begun, db.committed, db.rolledBack)
			}
			if tc.wantErr != nil {
				if got != (user.User{}) {
					t.Fatal("failed registration returned a user")
				}
				return
			}
			if got.ID != "user-id" || got.Name != "Alice" || got.Email != "alice@example.com" || db.userID != got.ID {
				t.Fatalf("unexpected user or account user ID: %+v, %q", got, db.userID)
			}
			if err := bcrypt.CompareHashAndPassword([]byte(db.hash), []byte(" password123 ")); err != nil {
				t.Fatalf("stored password hash: %v", err)
			}
		})
	}
}

func TestHashPasswordRejectsLongPassword(t *testing.T) {
	if _, err := hashPassword(strings.Repeat("a", 73)); !errors.Is(err, bcrypt.ErrPasswordTooLong) {
		t.Fatalf("error = %v, want ErrPasswordTooLong", err)
	}
}
