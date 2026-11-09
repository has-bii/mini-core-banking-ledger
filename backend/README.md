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

## Architecture

See [backend architecture guidelines](docs/rule.md) for layer responsibilities, transaction ownership, and validation rules.
