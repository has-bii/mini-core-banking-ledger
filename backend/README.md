# Mini Core Banking Ledger — Backend

A Go HTTP API for user registration, cookie-based sessions, and wallet management. Registration creates a main wallet. Signed-in users can create savings wallets and sub-wallets, and list or retrieve their own wallets. Transfers and ledger entries are not implemented.

PostgreSQL stores users and wallets; Redis stores sessions. See the [OpenAPI specification](api/openapi.yaml) for routes, request and response schemas, and errors.

## Run locally

Run commands from `backend/`. You'll need Go (version in [`go.mod`](go.mod)), Docker with Compose, and the [Goose CLI](https://github.com/pressly/goose) on your `PATH`. Compose provides PostgreSQL 18 (required for `uuidv7()`) and Redis.

If you don't already have a `.env`, copy the example:

```sh
cp .env.example .env
```

Check that `DATABASE_URL`, `REDIS_URL`, and `GOOSE_DBSTRING` match your local services. Set `GOOSE_DRIVER=postgres` and point `GOOSE_DBSTRING` at the same database as `DATABASE_URL`.

```sh
docker compose up -d --wait
goose -env .env -dir migrations up
go run ./cmd/api
```

The API listens at `http://localhost:8080` by default; API routes are under `/api`. The API loads `.env` without overriding existing environment variables. It does not run migrations on startup.

### Local behavior

- **Sessions:** `SESSION_LIFETIME` defaults to `30m` when unset or empty. Use a positive Go duration in whole seconds (for example, `8h`); invalid values prevent startup. Changes take effect for new logins after restarting the API and don’t change existing sessions. Local Compose Redis is nonpersistent, so restarting Redis signs everyone out.
- **Browser origin:** `CORS_ORIGIN` accepts one exact origin and defaults to `http://localhost:3000`.
- **Rate limiting:** Each API process shares one rate limit across clients. `RATE_LIMIT_CAPACITY` defaults to a burst of 50 requests, and `RATE_LIMIT_REFILL_RATE` defaults to 10 tokens per second. Requests that reach an empty bucket receive HTTP `429`.

## Tests and checks

### Unit tests

```sh
go test -race ./...
```

Unit tests don't need PostgreSQL, Redis, or an environment file.

### Integration tests

Start the Compose services and make sure Goose is on your `PATH`. A running API server is not required.

```sh
docker compose up -d --wait postgres redis
```

Create `.env.test` in `backend/` with URLs for dedicated test backends. With the default Compose credentials:

```dotenv
DATABASE_URL="postgresql://user:password@127.0.0.1:5432/test"
REDIS_URL="redis://:password@127.0.0.1:6379/1"
```

Compose creates the `test` database when it initializes a fresh PostgreSQL volume. If you're using an existing volume, make sure `test` exists and the test role can create schemas in it. Redis database `1` keeps local test keys separate from app sessions in database `0`. The test helpers read these URLs from `.env.test` directly; they don't fall back to `.env` or shell variables.

```sh
go test -race -count=1 -tags=integration ./...
```

This runs both unit and integration tests.

### Build and vet

```sh
go build ./...
go vet ./...
go vet -tags=integration ./...
```
