# Mini Core Banking Ledger

Application-specific setup and configuration instructions live alongside each app.

- [Backend](backend/README.md) — Go HTTP API backed by PostgreSQL.
- [Design system](design/MINILEDGER-DESIGN-SYSTEM.md) — shared visual guidelines and assets.

Shared PostgreSQL infrastructure is defined in [`docker-compose.postgres.yaml`](docker-compose.postgres.yaml).

## Local Redis

Start Redis using [`docker-compose.redis.yaml`](docker-compose.redis.yaml):

```sh
docker compose -f docker-compose.redis.yaml up -d --wait
```

Connect at `127.0.0.1:6379` with password `password`. Set `REDIS_PASSWORD` before starting to override it. This setup is for local development only; all Redis data is lost on restart.

Stop Redis:

```sh
docker compose -f docker-compose.redis.yaml down
```
