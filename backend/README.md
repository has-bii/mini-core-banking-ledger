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

Unit tests use Go's standard library and handwritten fakes; no PostgreSQL, `.env`, or extra test dependencies are required. Tests live beside the code in `*_test.go` files and cover auth validation, login/current-user behavior, registration preflight checks, session token handling, and email/UUID identifiers. HTTP-to-PostgreSQL integration tests live in `internal/app/app_integration_test.go` behind the `integration` build tag.


```sh
go test -race ./...
go vet ./...
```

### Integration tests

Use a dedicated test database. The tests read `DATABASE_URL` directly from `backend/.env.test`, ignoring the shell's `DATABASE_URL`; they never load `.env`. Ensure `.env.test` points at the test database and its `GOOSE_DBSTRING` matches `DATABASE_URL`.

With PostgreSQL running, apply migrations and run from `backend/`:

```sh
goose -env .env.test up
go test -tags=integration -count=1 ./...
```

The suite exercises registration, duplicate registration, wrong passwords, login/session cookies, and authenticated/unauthenticated current-user requests through the real HTTP handler and PostgreSQL. It uses unique user emails and deletes only its own users afterward (accounts and sessions cascade). Missing configuration, an unavailable database, or unapplied migrations fail the integration run rather than silently skipping it. Interrupted runs may leave test rows behind. Normal `go test ./...` remains database-free.

## Login and sessions

`POST /auth/login` accepts an email and password:

```sh
curl -i -c cookies.txt http://localhost:8080/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email":"alice@example.com","password":"password123"}'
```

A successful login returns the user and sets a host-only `session` cookie with `HttpOnly`, `SameSite=Lax`, `Path=/`, and a 24-hour lifetime. `Secure` is enabled outside `dev` and `test`; staging and production must use HTTPS. The raw token is never returned in JSON or stored in the database: the existing `sessions.token` column holds its SHA-256 hash. Invalid credentials return `401`; invalid input returns `400`.

`GET /auth/me` returns the current user using the session cookie from login:

```sh
curl -i -b cookies.txt http://localhost:8080/auth/me
```

A successful response is `200` with the same `{"data": {...}}` user shape as login. Missing, empty, invalid, expired, or revoked sessions return `401` with error code `UNAUTHENTICATED`; unexpected lookup failures return a generic `500`. Responses use `Cache-Control: no-store` and never include the session token or password hash.

The session service supports token lookup (rejecting expired sessions) and revocation. Reusable authentication middleware, logout endpoints, automatic expired-row cleanup, and login rate limiting are not included yet. The app's existing cross-origin protection applies to login too.

## Architecture

See [backend architecture guidelines](docs/architecture.md) for layer responsibilities, transaction ownership, and validation rules.
