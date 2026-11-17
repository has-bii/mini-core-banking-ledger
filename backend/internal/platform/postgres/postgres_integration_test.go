//go:build integration

package postgres_test

import (
	"context"
	"errors"
	"testing"

	"backend/internal/platform/postgres"
	"backend/internal/testutil"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestOpen_Query(t *testing.T) {
	fixture := testutil.Postgres(t)
	pool, err := postgres.Open(t.Context(), fixture.Config().ConnConfig.ConnString())
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(pool.Close)

	var got int
	if err := pool.QueryRow(t.Context(), "SELECT 42").Scan(&got); err != nil {
		t.Fatalf("query opened pool: %v", err)
	}
	if got != 42 {
		t.Errorf("query result = %d, want 42", got)
	}
}

func TestOpen_CanceledContext(t *testing.T) {
	fixture := testutil.Postgres(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	pool, err := postgres.Open(ctx, fixture.Config().ConnConfig.ConnString())
	if pool != nil {
		t.Cleanup(pool.Close)
		t.Error("Open returned a pool on failure")
	}
	if !errors.Is(err, context.Canceled) {
		t.Errorf("Open error = %v, want context.Canceled", err)
	}
}

func TestInTx_Commit(t *testing.T) {
	pool := transactionPool(t)
	err := postgres.InTx(t.Context(), pool, func(tx pgx.Tx) error {
		_, err := tx.Exec(t.Context(), "INSERT INTO transaction_entries (value) VALUES (42)")
		return err
	})
	if err != nil {
		t.Fatalf("InTx: %v", err)
	}

	var got int
	if err := pool.QueryRow(t.Context(), "SELECT value FROM transaction_entries").Scan(&got); err != nil {
		t.Fatalf("read committed entry: %v", err)
	}
	if got != 42 {
		t.Errorf("committed value = %d, want 42", got)
	}
}

func TestInTx_CallbackErrorRollsBack(t *testing.T) {
	pool := transactionPool(t)
	wantErr := errors.New("callback failed")
	var callbackTx pgx.Tx
	err := postgres.InTx(t.Context(), pool, func(tx pgx.Tx) error {
		callbackTx = tx
		if _, err := tx.Exec(t.Context(), "INSERT INTO transaction_entries (value) VALUES (42)"); err != nil {
			t.Fatalf("insert before callback failure: %v", err)
		}
		return wantErr
	})
	if !errors.Is(err, wantErr) {
		t.Errorf("InTx error = %v, want callback error", err)
	}
	assertRolledBack(t, pool, callbackTx)
}

func TestInTx_CommitFailureRollsBack(t *testing.T) {
	pool := transactionPool(t)
	var callbackTx pgx.Tx
	err := postgres.InTx(t.Context(), pool, func(tx pgx.Tx) error {
		callbackTx = tx
		// The deferred constraint lets the writes succeed and rejects only the commit.
		if _, err := tx.Exec(t.Context(), "INSERT INTO transaction_entries (value) VALUES (42), (42)"); err != nil {
			t.Fatalf("insert before commit failure: %v", err)
		}
		return nil
	})
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		t.Errorf("InTx error = %v, want PostgreSQL constraint error", err)
	} else if pgErr.Code != "23505" {
		t.Errorf("PostgreSQL error code = %s, want 23505 (unique violation)", pgErr.Code)
	}
	assertRolledBack(t, pool, callbackTx)
}

func TestInTx_BeginFailure(t *testing.T) {
	pool := testutil.Postgres(t)
	pool.Close()

	called := false
	err := postgres.InTx(t.Context(), pool, func(tx pgx.Tx) error {
		called = true
		return nil
	})
	if err == nil {
		t.Error("InTx on a closed pool succeeded, want begin failure")
	}
	if called {
		t.Error("InTx invoked callback after begin failure")
	}
}

func TestInTx_PanicRollsBack(t *testing.T) {
	pool := transactionPool(t)
	wantPanic := errors.New("callback panicked")
	var callbackTx pgx.Tx
	gotPanic := func() (recovered any) {
		defer func() { recovered = recover() }()
		if err := postgres.InTx(t.Context(), pool, func(tx pgx.Tx) error {
			callbackTx = tx
			if _, err := tx.Exec(t.Context(), "INSERT INTO transaction_entries (value) VALUES (42)"); err != nil {
				t.Fatalf("insert before panic: %v", err)
			}
			panic(wantPanic)
		}); err != nil {
			t.Fatalf("InTx: %v", err)
		}
		return nil
	}()
	if gotPanic != wantPanic {
		t.Errorf("recovered panic = %v, want callback panic", gotPanic)
	}
	assertRolledBack(t, pool, callbackTx)
}

func transactionPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	pool := testutil.Postgres(t)
	if _, err := pool.Exec(t.Context(), `CREATE TABLE transaction_entries (
		value integer UNIQUE DEFERRABLE INITIALLY DEFERRED
	)`); err != nil {
		t.Fatalf("create transaction test table: %v", err)
	}
	return pool
}

func assertRolledBack(t *testing.T, pool *pgxpool.Pool, tx pgx.Tx) {
	t.Helper()
	if tx == nil {
		t.Fatal("transaction callback was not invoked")
	}
	if _, err := tx.Exec(t.Context(), "SELECT 1"); !errors.Is(err, pgx.ErrTxClosed) {
		t.Fatalf("transaction after rollback error = %v, want pgx.ErrTxClosed", err)
	}
	var count int
	if err := pool.QueryRow(t.Context(), "SELECT count(*) FROM transaction_entries").Scan(&count); err != nil {
		t.Fatalf("read entries after rollback: %v", err)
	}
	if count != 0 {
		t.Errorf("entries after rollback = %d, want 0", count)
	}
}
