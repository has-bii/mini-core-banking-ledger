# Backend

Go HTTP API backed by PostgreSQL, using pgx and Goose SQL migrations.

Run all commands from `backend/`.

## Local setup

Requires Go matching `go.mod`, Docker Compose, and the [Goose CLI](https://github.com/pressly/goose) on `PATH`. PostgreSQL 18 is required for `uuidv7()`.

```sh
docker compose -f ../docker-compose.postgres.yaml up -d --wait
cp .env.example .env
goose -env .env up
go run ./cmd/api
```

The API loads `.env`; existing environment variables take precedence.

## Migrations

Keep `DATABASE_URL` and `GOOSE_DBSTRING` pointed at the same database. See `.env.example` for the required settings.

```sh
goose -env .env status
goose -env .env up
goose -env .env down # Rolls back the latest migration; may delete data.
goose -env .env create add_accounts sql
```

Edit generated files in `migrations/` under `-- +goose Up` and `-- +goose Down`. To migrate a test database, use `-env .env.test` instead.

## Testing

Unit tests need no database:

```sh
go test -race ./...
go vet ./...
```

For integration tests, copy `.env.example` to `.env.test` and point both database URLs at a separate, existing PostgreSQL 18 test database, never a default or production database. Compose creates a `test` database on first initialization. Its user must be allowed to create schemas.

```sh
go test -race -tags=integration -count=1 ./...
go vet -tags=integration ./...
```

Integration tests read `DATABASE_URL` only from `.env.test`, ignoring the shell value. They create, migrate, and drop an isolated schema per test; no manual migration or Goose CLI is needed. Configuration or database errors fail the run. A killed process may leave `test_*` schemas behind.

## Auth API

| Endpoint | Behavior |
| --- | --- |
| `POST /api/auth/login` | Accepts email and password; returns the user and sets a session cookie. |
| `GET /api/auth/me` | Returns the current user, or `401` for an invalid session. |
| `POST /api/auth/logout` | Revokes the current session and clears its cookie; returns `204`, even without a valid session. |

Sessions last 24 hours. Cookies use `HttpOnly` and `SameSite=Lax`; environments other than `dev` and `test` require HTTPS. Session tokens are stored as SHA-256 hashes and never returned in JSON. Login rate limiting and automatic expired-session cleanup are not implemented.

See [architecture guidelines](docs/architecture.md) for layer responsibilities, transactions, and validation rules.
