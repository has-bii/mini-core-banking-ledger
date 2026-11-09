package user

import (
	"backend/internal/platform/database"
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type testDB struct {
	database.DBTX
	row   pgx.Row
	ctx   context.Context
	query string
	args  []any
}

func (db *testDB) QueryRow(ctx context.Context, query string, args ...any) pgx.Row {
	db.ctx, db.query, db.args = ctx, query, args
	return db.row
}

type testRow func(...any) error

func (r testRow) Scan(dest ...any) error { return r(dest...) }

func TestStoreFind(t *testing.T) {
	dbErr := errors.New("database unavailable")
	image := "https://example.com/avatar.png"
	for _, lookup := range []struct {
		name  string
		find  func(*Store, context.Context, string) (User, error)
		value string
		query string
	}{
		{"by ID", (*Store).FindByID, "user-id", "SELECT * FROM users WHERE id = $1"},
		{"by email", (*Store).FindByEmail, "user@example.com", "SELECT * FROM users WHERE email = $1"},
	} {
		t.Run(lookup.name, func(t *testing.T) {
			for _, tc := range []struct {
				name    string
				image   *string
				scanErr error
				wantErr error
			}{
				{"found", &image, nil, nil},
				{"null image", nil, nil, nil},
				{"not found", nil, pgx.ErrNoRows, ErrNotFound},
				{"database error", nil, dbErr, dbErr},
			} {
				t.Run(tc.name, func(t *testing.T) {
					want := User{
						ID:        "user-id",
						Name:      "Test User",
						Email:     "user@example.com",
						Image:     tc.image,
						CreatedAt: time.Date(2026, time.January, 1, 0, 0, 0, 0, time.UTC),
						UpdatedAt: time.Date(2026, time.January, 2, 0, 0, 0, 0, time.UTC),
					}
					db := &testDB{row: testRow(func(dest ...any) error {
						if tc.scanErr != nil {
							return tc.scanErr
						}
						*dest[0].(*string) = want.ID
						*dest[1].(*string) = want.Name
						*dest[2].(*string) = want.Email
						*dest[3].(**string) = want.Image
						*dest[4].(*time.Time) = want.CreatedAt
						*dest[5].(*time.Time) = want.UpdatedAt
						return nil
					})}
					ctx, cancel := context.WithCancel(context.Background())
					defer cancel()
					got, err := lookup.find(NewStore(db), ctx, lookup.value)
					if !errors.Is(err, tc.wantErr) {
						t.Fatalf("error = %v, want %v", err, tc.wantErr)
					}
					if db.ctx != ctx || db.query != lookup.query || !reflect.DeepEqual(db.args, []any{lookup.value}) {
						t.Fatalf("unexpected query: context = %v, SQL = %q, args = %v", db.ctx, db.query, db.args)
					}
					if tc.wantErr != nil {
						want = User{}
					}
					if !reflect.DeepEqual(got, want) {
						t.Fatalf("user = %+v, want %+v", got, want)
					}
				})
			}
		})
	}
}

func TestStoreCreate(t *testing.T) {
	dbErr := errors.New("database unavailable")
	otherUniqueErr := &pgconn.PgError{Code: uniqueViolation, ConstraintName: "users_pkey"}
	for _, tc := range []struct {
		name             string
		userErr, wantErr error
	}{
		{name: "success"},
		{name: "user failure", userErr: dbErr, wantErr: dbErr},
		{name: "duplicate email", userErr: &pgconn.PgError{Code: uniqueViolation, ConstraintName: "users_email_key"}, wantErr: ErrEmailTaken},
		{name: "other unique constraint", userErr: otherUniqueErr, wantErr: otherUniqueErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := &testDB{row: testRow(func(dest ...any) error {
				if tc.userErr != nil {
					return tc.userErr
				}
				*dest[0].(*string) = "user-id"
				*dest[1].(*string) = "Alice"
				*dest[2].(*string) = "alice@example.com"
				return nil
			})}
			got, err := NewStore(db).Create(t.Context(), "Alice", "alice@example.com")
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("error = %v, want %v", err, tc.wantErr)
			}
			if tc.wantErr != nil {
				if got != (User{}) {
					t.Fatal("failed transaction returned a user")
				}
				return
			}
			if got.ID != "user-id" || got.Name != "Alice" || got.Email != "alice@example.com" {
				t.Fatalf("unexpected user: %+v", got)
			}
			if !reflect.DeepEqual(db.args, []any{"Alice", "alice@example.com"}) {
				t.Fatal("unexpected user insert arguments")
			}
		})
	}
}
