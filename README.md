# Mini Core Banking Ledger

Application-specific setup and configuration instructions live alongside each app.

- [Backend](backend/README.md) — Go HTTP API backed by PostgreSQL.
- [Design system](design/MINILEDGER-DESIGN-SYSTEM.md) — shared visual guidelines and assets.

Shared PostgreSQL and Redis infrastructure is defined in [`docker-compose.yaml`](docker-compose.yaml).

## Local infrastructure

Start PostgreSQL and Redis:

```sh
docker compose up -d --wait
```

Connect at `127.0.0.1:6379` with password `password`. Set `REDIS_PASSWORD` before starting to override it. This setup is for local development only; all Redis data is lost on restart.

To start only one service, use `docker compose up -d --wait postgres` or `docker compose up -d --wait redis`.

Stop both services (PostgreSQL data is preserved in its named volume):

```sh
docker compose down
```
