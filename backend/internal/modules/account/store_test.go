package account

import (
	"backend/internal/platform/database"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
)

type testDB struct {
	database.DBTX
	row    pgx.Row
	userID string
}

func (db *testDB) QueryRow(_ context.Context, _ string, args ...any) pgx.Row {
	db.userID = args[0].(string)
	return db.row
}

type testRow func(...any) error

func (r testRow) Scan(dest ...any) error { return r(dest...) }

func TestFindByUserID(t *testing.T) {
	dbErr := errors.New("database unavailable")
	for _, tc := range []struct {
		name    string
		scanErr error
		wantErr error
	}{
		{"found", nil, nil},
		{"not found", pgx.ErrNoRows, ErrNotFound},
		{"database error", dbErr, dbErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			now := time.Now()
			db := &testDB{row: testRow(func(dest ...any) error {
				if tc.scanErr != nil {
					return tc.scanErr
				}
				*dest[0].(*string) = "account-id"
				*dest[1].(*string) = "user-id"
				*dest[2].(**string) = nil
				*dest[3].(*time.Time) = now
				*dest[4].(*time.Time) = now
				return nil
			})}
			got, err := NewStore(db).FindByUserID(context.Background(), "user-id")
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("error = %v, want %v", err, tc.wantErr)
			}
			if db.userID != "user-id" {
				t.Fatalf("query user ID = %q", db.userID)
			}
			if err == nil && (got.ID != "account-id" || got.UserID != "user-id" || got.Password != nil || !got.CreatedAt.Equal(now) || !got.UpdatedAt.Equal(now)) {
				t.Fatalf("unexpected account: %+v", got)
			}
		})
	}
}
