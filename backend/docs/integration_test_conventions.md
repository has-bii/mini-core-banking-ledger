# Integration Test Conventions

## Roles

- **Store tests.** Use `store_integration_test.go` to verify persistence contracts against real backends.
- **App tests.** Use `internal/app/app_integration_test.go` when the assertion depends on assembled application wiring, such as registration followed by login and an authenticated request. Assert public responses or workflow outcomes.

## How to write

1. **Declare the test.** Put `//go:build integration` on the first line, then a blank line and the package declaration. Prefer an external package and standard `testing`; name store cases `TestStore_<MethodOrBehavior>`.
2. **Isolate the case.** Own the records you arrange and use a fixture per parallel case. Avoid seeded IDs, shared mutable state, and execution-order dependencies. Use named subtests and keep each table focused; prefer at most 12 cases per store behavior, splitting by behavior or explaining a larger table rather than dropping critical cases. Tests that use `t.Setenv` must run sequentially with no parallel ancestors.
3. **Run and assert.** Call the real constructor and operation with `t.Context()`. Pass the current `*testing.T` to helpers and call `t.Helper()` there. Register owned cleanup with `t.Cleanup`; give cleanup that outlives `t.Context()` its own bounded context.

## Errors

- **Prerequisite failures.** Stop with `t.Fatal` when a prerequisite fails. Missing configuration, services, tooling, or permissions fail the requested test rather than skipping it.
- **Independent checks.** Use `t.Error` when the other checks can still run.
- **Backend failure coverage.** Keep backend-specific failure matrices with the store. Follow the [store checks](store_conventions.md#how-to-test) for rollback, conflicts, and backend-specific outcomes.

## Store fixtures

Follow the [README integration setup](../README.md#integration-tests) for the test services, Goose, and `.env.test`. Fixtures read dedicated test URLs from that file rather than `.env` or shell fallbacks. Use test-only backends.

### PostgreSQL

Use `testutil.Postgres(t)` to apply the Goose migrations in a private schema. Keep queries on the fixture pool, whose `search_path` has no `public` fallback. The fixture closes the pool and drops only its own schema. Keep setup and cleanup scoped to that schema; never truncate or drop shared state.

Arrange a row independently of the lookup, then compare the returned fields with the inputs. This excerpt goes in `user/store_integration_test.go`:

```go
ctx := t.Context()
db := testutil.Postgres(t)
store := user.NewStore(db)
id := uuid.NewString()
const email = "alice@example.com"
if _, err := db.Exec(ctx, `INSERT INTO users (id, name, email, password) VALUES ($1, $2, $3, $4)`,
	id, "Alice", email, "stored-password-hash"); err != nil {
	t.Fatal(err)
}

got, err := store.GetByID(ctx, id)
if err != nil || got.ID != id || got.Email != email {
	t.Fatalf("GetByID() = (%q, %q), %v; want (%q, %q), nil", got.ID, got.Email, err, id, email)
}
```

### Redis

Use `testutil.Redis(t)` for an isolated key namespace, even when using a separate Redis database. Use the fixture client for setup and assertions so commands stay in that namespace. Cleanup deletes only the fixture's keys; never use `FLUSHDB` or `FLUSHALL`.

After writing a session through the store, check its TTL through the same fixture client. This excerpt goes in `session/store_integration_test.go`:

```go
ctx := t.Context()
client := testutil.Redis(t)
store := session.NewStore(client)
hash := strings.Repeat("a", 64)
expiresAt := time.Now().Add(time.Hour)
if _, err := store.Create(ctx, uuid.NewString(), hash, expiresAt); err != nil {
	t.Fatal(err)
}

ttl, err := client.PTTL(ctx, "session:"+hash).Result()
if err != nil || ttl <= 0 || ttl > time.Hour {
	t.Fatalf("session TTL = %v, %v; want a positive TTL of at most one hour", ttl, err)
}
```

## App wiring

Supply both fixtures to `app.New` and serve requests in-process with `httptest`.

## How to run

Run from `backend/`. The first command re-runs only the named Redis contract; the second includes unit tests as well as tagged tests. Use `-shuffle=on` when checking order independence. Report any checks you couldn't run and why.

```sh
go test -race -count=1 -tags=integration ./internal/modules/session -run '^TestStore_SessionLifecycle$'
go test -race -count=1 -tags=integration ./...
```
