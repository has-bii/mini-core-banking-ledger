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

Edit generated files in `migrations/` under `-- +goose Up` and `-- +goose Down`.

## Tests and checks

Run commands from `backend/`.

### Unit tests

```sh
go test -race ./...
```

Unit tests cover validation, HTTP parsing, distinct middleware contracts, and difficult-to-trigger failure handling. They use standard Go testing and `httptest`; small handwritten stubs are reserved for failure injection (middleware lookup callbacks are also permitted). No database, Redis, or environment file is required.

### Integration tests

Requires PostgreSQL 18 and the Goose CLI. Start the existing Compose database:

```sh
docker compose -f ../docker-compose.postgres.yaml up -d --wait
```

Create a local `.env.test` containing a dedicated test database URL:

```dotenv
DATABASE_URL="postgresql://user:password@localhost:5432/test"
```

Compose creates the `test` database when initializing a fresh volume. For an existing volume, ensure that database exists. `.env.test` is ignored by Git; do not use a production or development database.

```sh
go test -race -count=1 -tags=integration ./...
```

This runs both unit and integration tests. Integration tests use real services/stores for auth/session business outcomes, password/token hashing, persistence, expiry, rollback, and HTTP response/cookie contracts. The app test owns the wired register → login → me → logout flow without repeating lower-level input matrices or detailed storage checks. See the [testing guidelines](docs/testing_guideline.md) for contract ownership and permitted doubles. Redis and a running API server are not required.

The shared helper reads `DATABASE_URL` directly from `.env.test`, without falling back to `.env` or the shell's `DATABASE_URL`. It creates a unique schema per integration test, uses only that schema in each pool connection's `search_path`, and applies actual migrations using Goose. Cleanup closes the pool and drops only the schema it created; no shared tables are truncated. The database user needs schema creation privileges. Missing configuration, Goose, or PostgreSQL fails the tests rather than skipping them.

If a test process is forcibly killed, cleanup cannot run and a `test_…` schema may remain in the test database. Remove leftover schemas only after confirming no tests are using them.

### Compilation and static analysis

```sh
go build ./...
go vet ./...
go vet -tags=integration ./...
```

No coverage threshold is enforced.

## API specification

See [api/openapi.yaml](api/openapi.yaml) (OpenAPI 3.0.3) for implemented endpoints, request and response schemas, validation, errors, and session cookies.

See [architecture guidelines](docs/architecture.md) for layer responsibilities, transactions, and validation rules.
