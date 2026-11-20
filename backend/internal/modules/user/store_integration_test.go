//go:build integration

package user_test

import (
	"backend/internal/modules/user"
	"backend/internal/testutil"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

func TestStore_Create(t *testing.T) {
	t.Parallel()
	db := testutil.Postgres(t)
	store := user.NewStore(db)
	created, err := store.Create(t.Context(), "Alice", "alice@example.com", "stored-password-hash")
	if err != nil {
		t.Fatal(err)
	}
	id, err := uuid.Parse(created.ID)
	if err != nil || id.Version() != 7 || created.Name != "Alice" || created.Email != "alice@example.com" || created.Password != "stored-password-hash" || created.Image != nil || created.CreatedAt.IsZero() || created.UpdatedAt.IsZero() {
		t.Fatal("Create() did not return the persisted user fields and defaults")
	}
	var persisted user.User
	if err := db.QueryRow(t.Context(), `SELECT id, name, email, password, image, created_at, updated_at FROM users WHERE id = $1`, created.ID).Scan(
		&persisted.ID, &persisted.Name, &persisted.Email, &persisted.Password, &persisted.Image, &persisted.CreatedAt, &persisted.UpdatedAt,
	); err != nil {
		t.Fatal(err)
	}
	if persisted != created {
		t.Error("Create() result differs from persisted values")
	}
	var count int
	if err := db.QueryRow(t.Context(), `SELECT count(*) FROM wallets WHERE user_id = $1`, created.ID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Fatalf("wallet count = %d, want 1", count)
	}
	var name, walletType string
	var sortOrder, balance int64
	if err := db.QueryRow(t.Context(), `SELECT name, type, sort_order, balance FROM wallets WHERE user_id = $1`, created.ID).Scan(&name, &walletType, &sortOrder, &balance); err != nil {
		t.Fatal(err)
	}
	if name != "Main Wallet" || walletType != "main" || sortOrder != 0 || balance != 0 {
		t.Fatalf("main wallet = (%q, %q, %d, %d), want (Main Wallet, main, 0, 0)", name, walletType, sortOrder, balance)
	}
}

func TestStore_GetByIDAndEmail(t *testing.T) {
	t.Parallel()
	db := testutil.Postgres(t)
	store := user.NewStore(db)
	image := "https://example.com/avatar.png"
	want := user.User{
		ID: uuid.NewString(), Name: "Alice", Email: "alice@example.com", Password: "stored-password-hash", Image: &image,
		CreatedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC),
		UpdatedAt: time.Date(2026, 2, 3, 4, 5, 6, 0, time.UTC),
	}
	if _, err := db.Exec(t.Context(), `INSERT INTO users (id, name, email, password, image, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		want.ID, want.Name, want.Email, want.Password, want.Image, want.CreatedAt, want.UpdatedAt); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name string
		get  func(context.Context, string) (user.User, error)
		key  string
	}{
		{"by ID", store.GetByID, want.ID},
		{"by email", store.GetByEmail, want.Email},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.get(t.Context(), tt.key)
			if err != nil {
				t.Fatal(err)
			}
			if got.ID != want.ID || got.Name != want.Name || got.Email != want.Email || got.Password != want.Password ||
				got.Image == nil || *got.Image != image || !got.CreatedAt.Equal(want.CreatedAt) || !got.UpdatedAt.Equal(want.UpdatedAt) {
				t.Error("lookup did not map all persisted fields")
			}
		})
	}
}

func TestStore_MissingRecords(t *testing.T) {
	t.Parallel()
	store := user.NewStore(testutil.Postgres(t))
	for _, tt := range []struct {
		name string
		get  func(context.Context, string) (user.User, error)
		key  string
	}{
		{"missing ID", store.GetByID, uuid.NewString()},
		{"zero UUID is valid but absent", store.GetByID, uuid.Nil.String()},
		{"missing email", store.GetByEmail, "missing@example.com"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.get(t.Context(), tt.key)
			if got != (user.User{}) || !errors.Is(err, user.ErrNotFound) || errors.Is(err, pgx.ErrNoRows) {
				t.Fatalf("lookup error = %v, want zero user and domain ErrNotFound", err)
			}
		})
	}
}

func TestStore_GetByID_InvalidID(t *testing.T) {
	t.Parallel()
	db := testutil.Postgres(t)
	store := user.NewStore(db)
	db.Close() // Validation must happen before accessing even an unavailable database.
	for _, id := range []string{"", " ", "not-a-uuid", "019a3de4-8750-7000-8000-00000000000z"} {
		t.Run(id, func(t *testing.T) {
			got, err := store.GetByID(t.Context(), id)
			if got != (user.User{}) || !errors.Is(err, user.ErrInvalidUserID) {
				t.Fatalf("GetByID() error = %v, want zero user and ErrInvalidUserID", err)
			}
		})
	}
}

func TestStore_Create_UniqueConstraints(t *testing.T) {
	t.Parallel()
	db := testutil.Postgres(t)
	store := user.NewStore(db)
	if _, err := db.Exec(t.Context(), `ALTER TABLE users ADD CONSTRAINT test_users_name_key UNIQUE (name)`); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Create(t.Context(), "Alice", "alice@example.com", "stored-password-hash"); err != nil {
		t.Fatal(err)
	}
	got, err := store.Create(t.Context(), "Another Alice", "alice@example.com", "stored-password-hash")
	if got != (user.User{}) || !errors.Is(err, user.ErrEmailTaken) {
		t.Fatalf("duplicate email error = %v, want zero user and ErrEmailTaken", err)
	}
	got, err = store.Create(t.Context(), "Alice", "other@example.com", "stored-password-hash")
	var pgErr *pgconn.PgError
	if got != (user.User{}) || errors.Is(err, user.ErrEmailTaken) || !errors.As(err, &pgErr) || pgErr.Code != "23505" || pgErr.ConstraintName != "test_users_name_key" {
		t.Fatalf("unrelated unique constraint error = %v, want preserved PostgreSQL cause", err)
	}
}

func TestStore_UnexpectedFailures(t *testing.T) {
	t.Parallel()
	for _, failure := range []string{"missing table", "canceled context", "scan failure"} {
		t.Run(failure, func(t *testing.T) {
			t.Parallel()
			db := testutil.Postgres(t)
			store := user.NewStore(db)
			created, err := store.Create(t.Context(), "Alice", "alice@example.com", "stored-password-hash")
			if err != nil {
				t.Fatal(err)
			}
			ctx := t.Context()
			switch failure {
			case "missing table":
				if _, err := db.Exec(ctx, `DROP TABLE users CASCADE`); err != nil {
					t.Fatal(err)
				}
			case "scan failure":
				if _, err := db.Exec(ctx, `ALTER TABLE users ALTER COLUMN created_at DROP DEFAULT, ALTER COLUMN created_at TYPE text USING created_at::text, ALTER COLUMN created_at SET DEFAULT 'not-a-time'`); err != nil {
					t.Fatal(err)
				}
				db.Reset() // Discard cached result types so the operation reaches row scanning.
			case "canceled context":
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}
			for _, tt := range []struct {
				name string
				run  func(context.Context) (user.User, error)
			}{
				{"Create", func(ctx context.Context) (user.User, error) {
					return store.Create(ctx, "Bob", "bob@example.com", "stored-password-hash")
				}},
				{"GetByID", func(ctx context.Context) (user.User, error) { return store.GetByID(ctx, created.ID) }},
				{"GetByEmail", func(ctx context.Context) (user.User, error) { return store.GetByEmail(ctx, created.Email) }},
			} {
				t.Run(tt.name, func(t *testing.T) {
					got, err := tt.run(ctx)
					if got != (user.User{}) || err == nil {
						t.Fatal("failed operation must return no user and an error")
					}
					switch failure {
					case "missing table":
						var pgErr *pgconn.PgError
						if !errors.As(err, &pgErr) || pgErr.Code != "42P01" {
							t.Fatalf("error = %v, want missing-table cause", err)
						}
					case "scan failure":
						var scanErr pgx.ScanArgError
						if !errors.As(err, &scanErr) || scanErr.Err == nil || !errors.Is(err, scanErr.Err) {
							t.Fatalf("error = %v, want preserved scan cause", err)
						}
					case "canceled context":
						if !errors.Is(err, context.Canceled) {
							t.Fatalf("error = %v, want context.Canceled", err)
						}
					}
				})
			}
		})
	}
}

func TestStore_Create_MainWalletFailureRollsBack(t *testing.T) {
	t.Parallel()
	db := testutil.Postgres(t)
	if _, err := db.Exec(t.Context(), `ALTER TABLE wallets ADD CONSTRAINT test_reject_wallet CHECK (false)`); err != nil {
		t.Fatal(err)
	}
	got, err := user.NewStore(db).Create(t.Context(), "Alice", "alice@example.com", "stored-password-hash")
	var pgErr *pgconn.PgError
	if got != (user.User{}) || !errors.As(err, &pgErr) || pgErr.Code != "23514" || pgErr.ConstraintName != "test_reject_wallet" {
		t.Fatalf("Create() error = %v, want zero user and wallet constraint failure", err)
	}
	var users, wallets int
	if err := db.QueryRow(t.Context(), `SELECT (SELECT count(*) FROM users), (SELECT count(*) FROM wallets)`).Scan(&users, &wallets); err != nil {
		t.Fatal(err)
	}
	if users != 0 || wallets != 0 {
		t.Fatalf("failed creation left %d users and %d wallets", users, wallets)
	}
}

func TestStore_Create_ConcurrentEmailConflict(t *testing.T) {
	t.Parallel()
	db := testutil.Postgres(t)
	store := user.NewStore(db)
	type result struct {
		user user.User
		err  error
	}
	start := make(chan struct{})
	results := make(chan result, 2)
	for range 2 {
		go func() {
			<-start
			u, err := store.Create(t.Context(), "Alice", "alice@example.com", "stored-password-hash")
			results <- result{u, err}
		}()
	}
	close(start)
	var successes, conflicts int
	for range 2 {
		r := <-results
		switch {
		case r.err == nil && r.user.ID != "":
			successes++
		case errors.Is(r.err, user.ErrEmailTaken) && r.user == (user.User{}):
			conflicts++
		default:
			t.Errorf("unexpected concurrent creation error: %v", r.err)
		}
	}
	if successes != 1 || conflicts != 1 {
		t.Errorf("concurrent creation = %d successes, %d conflicts; want one each", successes, conflicts)
	}
	var users, wallets int
	if err := db.QueryRow(t.Context(), `SELECT (SELECT count(*) FROM users), (SELECT count(*) FROM wallets)`).Scan(&users, &wallets); err != nil {
		t.Fatal(err)
	}
	if users != 1 || wallets != 1 {
		t.Errorf("committed %d users and %d wallets, want one each", users, wallets)
	}
}
