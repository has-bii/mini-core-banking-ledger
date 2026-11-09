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

## API rate limiting

All `/api/` routes share one in-memory token bucket using `golang.org/x/time/rate`.
`RATE_LIMIT_RPS` defaults to 100 requests/second; `RATE_LIMIT_BURST` defaults to
200 tokens. Both must be positive integers. The bucket starts full, consumes one
token per request, and rejects excess requests immediately with HTTP 429 and
error code `RATE_LIMITED`. Allowed-origin CORS preflights bypass the bucket;
429 responses retain CORS headers.

Limits are per process and reset on restart. One client can consume the shared
allowance. These defaults need load testing; arrival limits do not cap concurrent
requests. No per-client or login-specific brute-force protection is provided.

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

## API specification

See [api/openapi.yaml](api/openapi.yaml) (OpenAPI 3.0.3) for implemented endpoints, request and response schemas, validation, errors, and session cookies.

See [architecture guidelines](docs/architecture.md) for layer responsibilities, transactions, and validation rules.
