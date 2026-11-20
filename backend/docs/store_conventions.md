# Store Conventions

## Roles

- **Persistence.** Take owner IDs explicitly and include the owner in the query predicate for reads and writes (`WHERE id = $1 AND user_id = $2`). Rely on database constraints for integrity, including uniqueness under concurrent writes.
- **Boundary.** Use already-validated DTOs as supplied and return domain values, not live database rows. `internal/app/app.go` supplies backend dependencies; keep connection setup out of the store and interfaces beside their consumers.

## How to write

1. **Wire the store.** In `store.go`, define a concrete `Store` and `NewStore` returning `*Store`. Split into `store_<concern>.go` only for distinct concerns. Accept `postgres.DBTX` for query-only PostgreSQL stores, `postgres.Beginner` when starting transactions, or the Redis client for Redis stores.
2. **Define the contract.** Name single lookups `GetBy<Field>` and collections `ListBy<Field>`. Use `Create`, `Update` (or `UpdateName`), and `Delete` (or `DeleteByTokenHash`) for writes. Put `ctx context.Context` first and pass it through. Use clear scalars or an already-validated DTO; document non-obvious preconditions, including non-nil pointer inputs. Document ordering and whether deletion is idempotent. Return zero items and nil error for an empty collection; its nilness isn't part of the contract.

### PostgreSQL

1. **Write the query.** Use parameterized pgx SQL local to the operation. List columns explicitly in scan order and share a private scanner only when mapping recurs. Use `RETURNING` for persisted or defaulted fields the caller needs. Close or consume rows (`pgx.CollectRows` can handle collections). Specify ordering when the contract requires it, including a tie-breaker if needed. Store money in integer minor units.
2. **Make multi-step writes atomic.** Use `postgres.InTx` for begin, commit, and rollback. Use its callback's `tx` for every participating query. For an owner-scoped sequence: lock the owner row with `SELECT ... FOR UPDATE`, read the last position, then insert and scan the new row.
3. **Keep the transaction bounded.** Use a single statement when it already provides atomicity. Keep hashing and external calls outside the transaction. Use an explicit boundary for a workflow that truly needs cross-store atomicity; a PostgreSQL transaction cannot include Redis writes.

### Redis

1. **Build and write the key.** Keep keys, commands, and serialization inside the store. Use a domain prefix such as `session:<tokenHash>` and persist only the caller-supplied opaque hash, never the raw session token. For sessions, validate the user ID as a UUID without querying PostgreSQL.
2. **Set lifetime atomically.** Derive TTL from the absolute expiration. Write the value and TTL together. Use `SET NX` with expiration when a collision must leave the original value untouched.
3. **Delete sessions idempotently.**

## Errors

- **Invalid identifiers.** Validate identifier arguments before accessing storage and return the corresponding sentinel, such as `ErrInvalidWalletID`.
- **Not found.** Translate `pgx.ErrNoRows` and `redis.Nil` to domain `ErrNotFound` with a zero model. Use the same result for another owner's record and for Redis records past their absolute expiration.
- **PostgreSQL conflicts and transactions.** Match recognized conflicts with `errors.As` to `*pgconn.PgError`, the SQLSTATE, and the constraint name. Return transaction callback errors to trigger rollback, and return a successful model only after commit succeeds.
- **Redis session writes and reads.** Reject a non-positive lifetime with `session.ErrInvalidExpiration` and report a `SET NX` collision as `session.ErrTokenHashTaken`. Treat bad JSON or missing required fields as corruption and preserve decoding causes where available.
- **Sentinels.** Define domain sentinels beside the module model and preserve them for `errors.Is`.
- **Unexpected failures.** Wrap with operation context using `%w`, including scan, iteration, collection, and decode failures. Return the error for the caller to handle, without logging it or including secrets. Return no successful model after a write failure; collections are usable only on nil error.

## How to test

1. **Set up `store_integration_test.go`.** Instantiate `NewStore` with a real backend fixture, not a mock. Use [integration test conventions](integration_test_conventions.md#store-fixtures) for the build tag, isolated PostgreSQL and Redis fixtures, and examples. Match split source files with `store_<concern>_integration_test.go`.
2. **Arrange state and call the store.** Create each case's own records and check the result against independent expected values. Assert persisted state, mapping, ordering, empty lists, identifier sentinels, ownership isolation, and relevant failures using `errors.Is` or `errors.As`.
3. **Check the persistence guarantee.** For atomic writes, inspect committed state and verify rollback after a later failure; coordinate concurrent calls explicitly when testing locks or constraints. For Redis, check value and TTL, collisions without overwrite, missing or corrupt records, and idempotent deletion. Force expiry with a past timestamp instead of sleeping.

Run the affected store integration tests from `backend/` (PostgreSQL and Goose for this example):

```sh
go test -race -count=1 -tags=integration ./internal/modules/wallet -run '^TestStore_'
```
