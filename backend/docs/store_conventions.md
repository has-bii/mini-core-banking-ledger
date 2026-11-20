# Store Conventions

The store is the module's persistence boundary. These conventions apply to new and changed PostgreSQL and Redis stores. Use [wallet/store.go](../internal/modules/wallet/store.go) as the main PostgreSQL example, not as a requirement to copy every implementation detail.

## Responsibilities

| Boundary | Owns | Does not own |
| --- | --- | --- |
| HTTP handler | Decoding, DTO normalization/validation, HTTP responses | SQL, Redis commands, persistence transactions |
| Service | Business eligibility, workflow orchestration, token generation/hashing | SQL, Redis commands, row scanning |
| Store | Queries/commands, mapping stored data, ownership predicates, persistence errors, atomic persistence operations | HTTP details, business workflows, repeated DTO validation |
| Platform | Connections and shared transaction lifecycle helpers | Module queries and domain rules |

A store may coordinate multiple queries, locks, or writes needed to implement one persistence operation. Wallet creation, for example, locks the owner and allocates the next sort order before inserting. Services decide whether the business operation is allowed; stores make its persistence safe.

Wire stores explicitly in `internal/app/app.go`. Stores receive their storage dependency; they do not open connections, read environment files, construct collaborating stores, or call services. Keep platform packages independent of modules.

## Files and dependencies

Keep persistence beside its module:

```text
internal/modules/wallet/
├── wallet.go                  # Model and domain sentinels
├── dto.go                     # Input DTOs and validation
├── store.go                   # Store, constructor, queries, scanning
├── store_integration_test.go  # All store tests
└── service.go                 # Consumer-defined store interface
```

- Use concrete `Store` and `NewStore`; the module name provides context (`wallet.Store`). Do not add a second persistence abstraction or generic CRUD framework.
- Keep small interfaces beside their consumers, such as `walletStore` in `service.go`. Expose only methods the consumer needs; constructors return `*Store`.
- PostgreSQL stores that begin transactions accept `postgres.Beginner`. Query-only stores accept `postgres.DBTX`, satisfied by both `*pgxpool.Pool` and `pgx.Tx`.
- Redis stores accept the Redis client, following `session.NewStore`. Do not add interfaces solely for tests.
- Start with `store.go`. Split by meaningful concern into `store_<concern>.go` only when useful; use matching `store_<concern>_integration_test.go` files.
- Keep helpers private and specific: `scanWallet`, not exported mapping utilities or an unrelated shared helper package.

## Method names and signatures

| Intent | Name | Typical result |
| --- | --- | --- |
| Fetch one record | `GetByID`, `GetByEmail`, `GetByTokenHash` | `(Model, error)` |
| Fetch a collection | `ListByUserID` | `([]Model, error)` |
| Insert | `Create` | `(Model, error)` |
| Change persisted fields | `Update` or a specific name such as `UpdateName` | `(Model, error)` or `error`, according to the contract |
| Remove | `Delete`, qualified when needed (`DeleteByTokenHash`) | `error` |

- Prefer `GetBy<Field>` for one result; do not introduce a separate `Find` versus `Get` semantic distinction.
- Avoid repeating the model name in methods: `wallet.Store.Create`, not `CreateWallet`.
- Put `ctx context.Context` first and pass it through to storage operations. Keep receiver names short (`s`). Use Go initialisms consistently: `userID`, `walletID`.
- Use explicit lookup and ownership arguments; for example, `GetByID(ctx, walletID, userID)` filters by both. Keep argument order consistent across the consumer interface and implementation.
- Use a DTO for structured write inputs when appropriate; a few scalar arguments remain reasonable. Do not invent a second input type solely to separate layers.
- Return module models or values, never HTTP responses or live database rows. Do not expose secrets in public JSON.
- Document non-obvious preconditions, ordering, and deletion semantics. Missing-record deletion may be idempotent, as with session revocation; do not leave callers guessing.

## Validation boundary

**Validate a DTO once at entry, not again in the store.**

- HTTP handlers normalize and validate DTOs before calling services. Non-HTTP callers must satisfy the same precondition before passing a DTO into the workflow. Pointer DTOs must be non-nil unless explicitly documented otherwise.
- Stores do not call `DTO.Validate()`, repeat DTO field rules, or silently normalize/mutate a DTO. `wallet.Store.Create` accepts an already-normalized, validated `CreateWalletDTO`.
- Stores still validate identifier arguments before accessing storage, returning module sentinels such as `ErrInvalidWalletID` and `ErrInvalidUserID`.
- Services enforce business policy and application permissions. Stores scope owned reads/writes in the query itself; a prior service check does not replace `WHERE id = $1 AND user_id = $2`.
- Database constraints remain authoritative for uniqueness, foreign keys, nullability, and other integrity rules. A DTO check or availability lookup cannot prevent a concurrent write from racing.
- Input validation returns structured `validator.ErrorValidation` where applicable. Preserve it with `%w` if it crosses a boundary; do not flatten field details into an error string. Stores need not generate DTO validation errors.

## Results and errors

Domain sentinels live beside the module's model. Recognized domain outcomes use those sentinels, directly or wrapped; callers use `errors.Is`, not message matching.

| Outcome | Contract |
| --- | --- |
| No matching single record | Zero model and domain `ErrNotFound` |
| Record belongs to another owner | Same `ErrNotFound`; do not disclose its existence |
| No matching collection | Zero items and nil error; callers must not depend on nil versus empty slice |
| Invalid identifier | Zero result and the corresponding domain sentinel |
| Recognized persistence conflict | Specific domain sentinel, such as `user.ErrEmailTaken` |
| Unexpected storage, scan, or decoding failure | Zero single-record result and an error retaining the original cause |

- Translate `pgx.ErrNoRows` and `redis.Nil` at the store boundary rather than leaking them as missing-record contracts.
- Match recognized PostgreSQL violations precisely using `errors.As` with `*pgconn.PgError`, SQLSTATE, and the relevant constraint name. Do not translate every unique violation into an email conflict.
- Wrap unexpected failures with operation context and `%w`, including row collection/scanning failures: `fmt.Errorf("list wallets: %w", err)`. Keep causes available to `errors.Is`/`errors.As`; do not replace all infrastructure failures with a generic sentinel.
- On write or commit failure, return no successful model. Collection results are usable only when the returned error is nil.
- Do not log and return the same failure at every layer. Let the application/transport boundary log unexpected errors; error messages must not embed passwords, tokens, or submitted secrets.

## PostgreSQL queries and transactions

- Use pgx and parameterized SQL (`$1`, `$2`); no ORM or query builder. Never interpolate user-supplied values or unchecked identifiers into SQL.
- List columns explicitly, keeping `SELECT`/`RETURNING` order aligned with scanning. Share a small scanner helper when the same mapping recurs.
- Keep SQL local to the operation (`const q` when helpful). Return persisted/defaulted values through `RETURNING` when the caller needs the created or updated record.
- Close/consume rows correctly; `pgx.CollectRows` is appropriate for collection mapping. Preserve query, scan, and iteration errors.
- Specify ordering when callers rely on it. Wallet lists use `ORDER BY sort_order ASC`; choose a tie-breaker when the primary order is not unique within the query scope.
- Store money exactly, using integer minor units as wallet balances do. Never use floating-point amounts.

**The store owns the transaction for its atomic persistence operation; `postgres.InTx` owns begin/commit/rollback.**

Wallet creation is the reference sequence:

1. Validate `userID`; accept the caller-validated DTO without revalidation.
2. Call `postgres.InTx(ctx, s.db, ...)`.
3. Lock the owner row with `SELECT ... FOR UPDATE` to serialize allocation for that owner.
4. Read the last wallet sort order through the transaction.
5. Insert with the next sort order, scanning the returned record.
6. Return the model only after `InTx` succeeds.

Use the callback's `tx` for every participating query; calling the pool would move work outside the transaction. Return callback failures so the helper rolls back. Keep expensive work such as password hashing and external calls outside the transaction. A single atomic statement needs no explicit transaction.

Do not implement manual transaction lifecycle in each store, hide nested transactions, or assume sequential calls to separate stores are atomic. If a future workflow requires cross-store atomicity, design an explicit shared transaction boundary for that workflow; do not add speculative transaction infrastructure now. PostgreSQL transactions cannot atomically include Redis operations.

## Redis persistence

Use [session/store.go](../internal/modules/session/store.go) as the Redis example. Keep key construction, commands, serialization, and decoding inside the store.

- Use domain-prefixed keys (`session:<tokenHash>`). Token generation and hashing belong to the service; never persist or log raw session tokens. Session store callers must supply a service-generated token hash; the store treats it as an opaque key without format validation or normalization.
- Session creation validates the user ID as a UUID and returns `session.ErrInvalidUserID` before Redis access. It does not query PostgreSQL to verify user existence.
- Give temporary records explicit expiry. Store the value and TTL atomically; use `SET NX` with expiration when collisions must not overwrite an existing record, as session creation does.
- Derive TTL from the absolute expiration; reject non-positive remaining lifetime with `session.ErrInvalidExpiration`. Report creation collisions as `session.ErrTokenHashTaken`. Do not create a value first and attach its TTL in a separate command.
- Treat missing or expired sessions as domain `ErrNotFound`. Check decoded absolute expiry as well as relying on Redis TTL.
- Treat malformed JSON or missing required stored fields as unexpected corruption, not as a missing session. Preserve decoding causes where available.
- Make session deletion idempotent. A successful delete of an absent key is not an error.
- Translate recognized domain failures into sentinels; preserve network/command failures with operation context. Do not promise SQL-style transactions or cross-backend atomicity.

## Integration testing

**Stores have integration tests only.** Test real persistence, not mocks, command stubs, or nil dependencies. DTO rule/normalization tests belong with the DTO; HTTP tests own payload rejection and status mapping; service tests own business rules. Do not repeat their input matrices in store tests.

### Structure and contracts

- Put `//go:build integration` first in `store_integration_test.go`. Prefer external packages such as `package wallet_test` and test exported behavior.
- Use standard `testing`; name tests `TestStore_<MethodOrBehavior>` and table subtests by their behavior. Do not add assertion libraries, generated mocks, or test-only production interfaces.
- Cover mapping and persisted values, empty lists, missing records, identifier sentinels, ownership isolation, recognized constraint translation, and unexpected failures relevant to the operation.
- For atomic operations, verify committed state, rollback when a later step fails, and concurrent behavior where locks/constraints protect integrity. Coordinate goroutines explicitly, not with sleeps.
- For Redis, cover value/TTL persistence, collisions without overwrites, missing/expired records, corruption, and idempotent deletion. Force past expiry timestamps instead of sleeping.
- Assert observable values and state plus error contracts (`errors.Is`/`errors.As`). Exact internal error strings are not contracts. Use independent expected values.
- Preserve useful coverage when moving tests: wallet's identifier cases belong in store integration tests; repeated DTO-validation assertions do not.
- Keep cases focused. Start with success, distinct failures, and relevant boundaries. Prefer at most 12 cases per behavior table; split by behavior or explain a necessary larger table. Never drop security, money, integrity, or regression cases to satisfy a count.

### Isolation and safety

- Use `testutil.Postgres(t)` for PostgreSQL 18 and real Goose migrations in a private schema. Use `testutil.Redis(t)` for Redis with a private key namespace.
- Fixtures read dedicated test URLs explicitly from ignored `backend/.env.test`, never `.env` or shell fallbacks. Never point them at production or development data. Missing configuration, services, Goose, or permissions fails requested integration tests rather than skipping them.
- Keep PostgreSQL `search_path` restricted to the test schema, with no `public` fallback. Close pools before dropping only that test's schema; never truncate shared tables or drop shared schemas/databases.
- Clean up only the Redis fixture's keys; never use `FLUSHDB` or `FLUSHALL`. A separate test database alone does not replace per-test namespacing.
- Create each test's own starting records; no seeded IDs, shared mutable fixtures, or order dependencies. Use `t.Parallel()` only with isolated state; tests using `t.Setenv` cannot be parallel or have parallel ancestors.
- Use `t.Context()`, `t.Cleanup`, and helpers calling `t.Helper()`. Cleanup that outlives the test context needs its own bounded context. Bind assertions to the current subtest's `*testing.T`.
- Fail prerequisites with `t.Fatal`/`t.Fatalf`; use `t.Error`/`t.Errorf` for independent checks. Never ignore fixture/setup errors.

### Verification

Run from `backend/`; see [README setup](../README.md#tests-and-checks):

```sh
go test -race -count=1 -tags=integration ./internal/modules/wallet # Affected store first
go test -race ./...                                            # Other unit contracts
go test -race -count=1 -tags=integration ./...                    # Full integration suite
go build ./...
go vet ./...
go vet -tags=integration ./...
```

Use `-shuffle=on` when checking order independence. Report checks not run and why; no coverage percentage or minimum test count is required.

## User, session, and auth alignment

- User lookups use `GetByID` and `GetByEmail`. The user store accepts `postgres.DBTX`, validates lookup UUIDs with `user.ErrInvalidUserID`, and accepts caller-validated scalar creation inputs with an already-hashed password.
- Session lookup uses `Store.GetByTokenHash` and `Service.GetByToken`. Creation validates user IDs and exposes the expiration/collision sentinels above; deletion remains idempotent.
- Auth consumes the user store and session service; it has no separate store. Registration and login require caller-normalized, validated DTOs. Password hashing and workflows stay in the service.
- `auth.Service.Me` receives a verified identity. A missing user maps to `auth.ErrUnauthenticated` (HTTP 401); a malformed user ID preserves `user.ErrInvalidUserID` as an internal contract failure (HTTP 500), not an authentication outcome.

These API and error-contract changes do not change registration/login/logout policy, session lifetime or cookies, schema, or automatic main-wallet creation.
