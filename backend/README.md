# Backend

Go HTTP API backed by PostgreSQL, using pgx and Goose SQL migrations.

All commands below run from `backend/`.

## Local setup

Requirements: Go matching `go.mod`, Docker Compose, and the [Goose CLI](https://github.com/pressly/goose) on `PATH`.

```sh
docker compose -f ../docker-compose.postgres.yaml up -d --wait
cp .env.example .env
goose -env .env up
go run ./cmd/api
```

The API loads `.env` from the working directory; existing environment variables take precedence. PostgreSQL 18 is required by the `uuidv7()` defaults in the migrations.

## Database migrations

Run Goose from `backend/`. Use `.env` for the default database and `.env.test` for the test database. Both files should define `GOOSE_DRIVER`, `GOOSE_DBSTRING`, and `GOOSE_MIGRATION_DIR` (see `.env.example`). Keep `DATABASE_URL` and `GOOSE_DBSTRING` pointed at the same database within each file.

### Default database (`.env`)

```sh
goose -env .env status
goose -env .env up
# Roll back the most recent migration (may delete data).
goose -env .env down
```

### Test database (`.env.test`)

If `.env.test` does not exist, copy `.env.example` to `.env.test` and change both database URLs to a separate, existing test database before running migrations. Do not point it at your default or production database.

```sh
goose -env .env.test status
goose -env .env.test up
# Roll back the most recent test migration.
goose -env .env.test down
```

### Create a migration

```sh
goose -env .env create add_accounts sql
```

Edit the generated file in `migrations/`, adding the schema changes under `-- +goose Up` and their rollback under `-- +goose Down`. Apply it to the appropriate database using the commands above.

## Testing

Database-free tests use Go's standard library and handwritten adapters. They cover auth validation and workflows, session token handling, identifiers, and HTTP routing, error responses, and cookie behavior. Auth and session tests use external test packages so they exercise the exported interface rather than private implementation details.

PostgreSQL-backed tests live beside the relevant modules in `*_integration_test.go`, behind the `integration` build tag. The full authentication journey lives in `internal/app/app_integration_test.go`.

```sh
go test -race ./...
go vet ./...
```

### Integration tests

Use a dedicated PostgreSQL 18 test database. The tests read `DATABASE_URL` directly from `backend/.env.test`, ignoring the shell's `DATABASE_URL`; they never load `.env`. The database must already exist and its user must be allowed to create schemas. The Compose setup creates a database named `test` on first initialization.

With PostgreSQL running, run from `backend/`:

```sh
go test -race -tags=integration -count=1 ./...
go vet -tags=integration ./...
# Check that scenarios do not depend on execution order.
go test -race -tags=integration -shuffle=on -count=3 ./...
```

`internal/testutil/testdb.Open(t)` creates a unique schema per test, sets the pool's search path to that schema, applies the checked-in SQL migrations' Goose Up sections in transactions, and drops only that schema during cleanup. No manual migration step or Goose executable is needed for tests. This helper supports SQL-only migrations, not Goose Go migrations or advanced Goose directives. It does not migrate or clear the public schema. Parallel tests can reuse fixture emails without sharing rows; normal cleanup also runs on test failure. A killed process may leave `test_*` schemas behind in the dedicated test database.

Coverage is split by behavior:

- **Auth module:** persisted registration credentials, rollback on credentials failure, and concurrent duplicate registrations that both pass the pre-check.
- **Stores:** user/account round-trips, domain-error translation, PostgreSQL constraint failures, session expiry, and scoped, idempotent revocation.
- **Application:** one register → login → me → logout journey, plus independent error, multi-session, and environment-specific cookie scenarios.

Missing configuration, an unavailable database, or migration failures fail the integration run rather than silently skipping it. Normal `go test ./...` remains database-free.

## Login and sessions

`POST /api/auth/login` accepts an email and password:

```sh
curl -i -c cookies.txt http://localhost:8080/api/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","password":"password123"}'
```

A successful login returns the user and sets a host-only `session` cookie with `HttpOnly`, `SameSite=Lax`, `Path=/`, and a 24-hour lifetime. `Secure` is enabled outside `dev` and `test`; staging and production must use HTTPS. The raw token is never returned in JSON or stored in the database: the existing `sessions.token` column holds its SHA-256 hash. Invalid credentials return `401`; invalid input returns `400`.

`GET /api/auth/me` returns the current user using the session cookie from login:

```sh
curl -i -b cookies.txt http://localhost:8080/api/auth/me
```

A successful response is `200` with the same `{"data": {...}}` user shape as login. Missing, empty, invalid, expired, or revoked sessions return `401` with error code `UNAUTHENTICATED`; unexpected lookup failures return a generic `500`. Responses use `Cache-Control: no-store` and never include the session token or password hash.

`POST /api/auth/logout` revokes only the current session and clears the session cookie:

```sh
curl -i -b cookies.txt -c cookies.txt -X POST http://localhost:8080/api/auth/logout
```

Logout returns `204 No Content`, including when the cookie is missing, empty, invalid, expired, or already revoked. Other sessions remain active. Unexpected revocation failures return a generic `500` without clearing the cookie so the client can retry. Responses use `Cache-Control: no-store`.

The session service supports token lookup (rejecting expired sessions) and revocation. Reusable authentication middleware, automatic expired-row cleanup, and login rate limiting are not included yet. The app's existing cross-origin protection applies to login and logout too.

## Architecture

See [backend architecture guidelines](docs/architecture.md) for layer responsibilities, transaction ownership, and validation rules.
