package user

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type credentialsDB struct {
	*testDB
	tx       *credentialsTx
	beginErr error
}

func (db *credentialsDB) Begin(context.Context) (pgx.Tx, error) { return db.tx, db.beginErr }

type credentialsTx struct {
	pgx.Tx
	userErr, accountErr, commitErr error
	userArgs, accountArgs          []any
	committed, rolledBack          bool
}

func (tx *credentialsTx) QueryRow(_ context.Context, _ string, args ...any) pgx.Row {
	tx.userArgs = args
	return testRow(func(dest ...any) error {
		if tx.userErr != nil {
			return tx.userErr
		}
		*dest[0].(*string) = "user-id"
		*dest[1].(*string) = args[0].(string)
		*dest[2].(*string) = args[1].(string)
		return nil
	})
}

func (tx *credentialsTx) Exec(_ context.Context, query string, args ...any) (pgconn.CommandTag, error) {
	if !strings.Contains(query, "INSERT INTO accounts") {
		panic("unexpected SQL: " + query)
	}
	tx.accountArgs = args
	return pgconn.CommandTag{}, tx.accountErr
}

func (tx *credentialsTx) Commit(context.Context) error {
	tx.committed = true
	return tx.commitErr
}

func (tx *credentialsTx) Rollback(context.Context) error {
	tx.rolledBack = true
	return nil
}

func TestCreateWithCredentials(t *testing.T) {
	dbErr := errors.New("database unavailable")
	for _, tc := range []struct {
		name                                              string
		beginErr, userErr, accountErr, commitErr, wantErr error
		wantCommit, wantUser, wantAccount                 bool
	}{
		{name: "success", wantCommit: true, wantUser: true, wantAccount: true},
		{name: "begin failure", beginErr: dbErr, wantErr: dbErr},
		{name: "user failure", userErr: dbErr, wantErr: dbErr, wantUser: true},
		{name: "duplicate email", userErr: &pgconn.PgError{Code: uniqueViolation, ConstraintName: "users_email_key"}, wantErr: ErrEmailTaken, wantUser: true},
		{name: "account failure", accountErr: dbErr, wantErr: dbErr, wantUser: true, wantAccount: true},
		{name: "commit failure", commitErr: dbErr, wantErr: dbErr, wantCommit: true, wantUser: true, wantAccount: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			tx := &credentialsTx{userErr: tc.userErr, accountErr: tc.accountErr, commitErr: tc.commitErr}
			// A separate transaction fake ensures writes never use the pool directly.
			db := &credentialsDB{tx: tx, beginErr: tc.beginErr}
			got, err := NewStore(db).CreateWithCredentials(t.Context(), "Alice", "alice@example.com", "hashed-password")
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("error = %v, want %v", err, tc.wantErr)
			}
			if tx.committed != tc.wantCommit || tx.rolledBack != (tc.beginErr == nil) {
				t.Fatalf("commit = %v, rollback cleanup = %v", tx.committed, tx.rolledBack)
			}
			if (tx.userArgs != nil) != tc.wantUser || (tx.accountArgs != nil) != tc.wantAccount {
				t.Fatalf("unexpected writes: user = %v, account = %v", tx.userArgs, tx.accountArgs)
			}
			if tc.wantErr != nil {
				if got != (User{}) {
					t.Fatal("failed registration returned a user")
				}
				return
			}
			if got.ID != "user-id" || got.Name != "Alice" || got.Email != "alice@example.com" ||
				!reflect.DeepEqual(tx.userArgs, []any{"Alice", "alice@example.com"}) ||
				!reflect.DeepEqual(tx.accountArgs, []any{"user-id", "hashed-password"}) {
				t.Fatalf("unexpected user or insert arguments: %+v, %v, %v", got, tx.userArgs, tx.accountArgs)
			}
		})
	}
}
