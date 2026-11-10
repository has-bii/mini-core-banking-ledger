# Backend

Go HTTP API backed by PostgreSQL, using pgx and Goose SQL migrations.

Run all commands from `backend/`.

## Local setup

Requires Go matching `go.mod`, Docker Compose, and the [Goose CLI](https://github.com/pressly/goose) on `PATH`. PostgreSQL 18 is required for `uuidv7()`.

```sh
docker compose up -d --wait
cp .env.example .env
goose -env .env up
go run ./cmd/api
```

The API loads `.env`; existing environment variables take precedence.

### Rate limiting

| Variable | Default | Meaning |
| --- | --- | --- |
| `RATE_LIMIT_CAPACITY` | `50` | Maximum tokens (burst size); must be at least `1` |
| `RATE_LIMIT_REFILL_RATE` | `10` | Tokens replenished per second; must be greater than `0` |

Both accept finite decimal values; invalid settings prevent startup. The default burst size was previously hard-coded to `20` and is now `50`; set `RATE_LIMIT_CAPACITY=20` to retain the previous limit.

Each API process has one in-memory token bucket shared by all clients and routes. It starts full, consumes one token per request, and returns HTTP `429` with error code `TOO_MANY_REQUESTS` when exhausted. CORS preflight requests bypass the bucket. Restarting resets it; separate instances do not share limits. This is not per-user or per-IP throttling.

## Migrations

Keep `DATABASE_URL` and `GOOSE_DBSTRING` pointed at the same database. See `.env.example` for the required settings.

```sh
goose -env .env status
goose -env .env up
goose -env .env down # Rolls back the latest migration; may delete data.
goose -env .env create add_user_field sql
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
docker compose up -d --wait postgres
```

Create a local `.env.test` containing a dedicated test database URL:

```dotenv
DATABASE_URL="postgresql://user:password@localhost:5432/test"
```

Compose creates the `test` database when initializing a fresh volume. For an existing volume, ensure that database exists. `.env.test` is ignored by Git; do not use a production or development database.

```sh
go test -race -count=1 -tags=integration ./...
```

This runs both unit and integration tests. Integration tests use real services/stores for auth/session business outcomes, password/token hashing, persistence, expiry, atomic registration, transaction rollback, and HTTP response/cookie contracts. The app test owns the wired register → login → me → logout flow without repeating lower-level input matrices or detailed storage checks. See the [testing guidelines](docs/testing_guideline.md) for contract ownership and permitted doubles. Redis and a running API server are not required.

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
