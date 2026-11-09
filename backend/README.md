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

## Architecture

See [backend architecture guidelines](docs/architecture.md) for layer responsibilities, transaction ownership, and validation rules.
