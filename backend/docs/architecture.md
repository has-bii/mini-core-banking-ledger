# Backend Architecture Guidelines

**Pattern:** Handler → Service → Store (Controller-Service-Repository)
**Stack:** Go, PostgreSQL, pgx; parameterized SQL, no ORM or query builder.

## 1. Layer Responsibilities

| Layer | Owns | Must not contain |
| --- | --- | --- |
| **Handler** | HTTP routing, decoding, payload validation, status codes, JSON responses | Business workflows, SQL, transactions |
| **Service** | Business rules, workflow orchestration, transaction boundaries | HTTP details, SQL |
| **Store** | Parameterized queries, row scanning, persistence error translation | Business workflows, HTTP details, transaction ownership |
| **PostgreSQL platform** | Pool setup and transaction lifecycle through `postgres.InTx` | Domain rules, application queries |

### Handlers

- Decode and validate request payloads before calling a service.
- Return errors before writing a response; register error-returning handlers through `httpx.Adapt`.
- Let the HTTP adapter map application codes to statuses, log unexpected errors once, and return safe JSON responses.

### Services

- Enforce business rules and coordinate stores.
- Decide which operations must succeed or fail together.
- Use `postgres.InTx` with a pool for atomic workflows; construct participating stores with the callback's `pgx.Tx`.
- Keep expensive work that does not need a transaction, such as password hashing, outside it.
- Return safe domain models or DTOs; never expose password hashes.
- Translate expected domain failures into `apperror.Error`, preserving the original cause. Propagate unexpected errors with useful `%w` context.

### Stores

- Accept `postgres.DBTX`, which both `*pgxpool.Pool` and `pgx.Tx` satisfy.
- Execute queries through the supplied dependency. The same store works inside or outside a transaction.
- Translate recognized persistence errors into domain errors, such as `user.ErrEmailTaken` or `user.ErrNotFound`.
- Wrap unexpected errors with operation context using `%w`.
- Do not begin transactions or construct other stores.

## 2. Dependencies and Wiring

`internal/app/app.go` is the composition root. Wire dependencies explicitly with constructors; no IoC container is required.

### Platform packages

```text
internal/platform/
├── apperror/                 # Client-safe application errors and stable public codes
├── transport/httpx/          # JSON decoding, responses, error adapter, and HTTP status mapping
│   └── middleware/           # Chaining, recovery, logging, CORS, auth, rate limiting
├── storage/
│   ├── postgres/             # Connection pool, DBTX, and transaction lifecycle
│   └── redis/                # Redis connection setup
├── observability/logger/     # Logger setup
└── validation/               # Field errors and UUID validation
```

Platform packages must not import `internal/modules`. Authentication middleware accepts a user-ID lookup function; `app.New` adapts the session service, translating missing or expired sessions into an empty user ID with no error. Other lookup errors remain errors and produce a generic `500` response. User-context helpers live beside authentication middleware.

### Wiring

- **Handlers** receive services and supporting infrastructure such as a logger.
- **Services** receive stores and collaborating services through consumer-defined interfaces. Construct pool-backed stores in `app.New`; additionally inject a pool when a service owns transactional workflows. Transaction-bound stores are constructed inside the transaction callback.
- **Stores** receive `postgres.DBTX`.
- Services may depend on other services for shared business logic, but dependencies must remain acyclic. Extract shared logic when needed.

```mermaid
flowchart TD
    App["app.New: composition root"] --> Handler["Auth handler"]
    App --> Service["Auth service"]
    App --> Pool["pgxpool.Pool"]
    Handler -->|"validated payload"| Service
    Service -->|"registration and identity"| UserStore["User store: pool"]
    Service --> SessionService["Session service"]
    SessionService --> SessionStore["Session store: pool"]
    Pool -.->|"supplied to stores"| UserStore
    Pool -.-> SessionStore
    UserStore --> DB[(PostgreSQL)]
    SessionStore --> DB
```

## 3. Transaction Ownership

**The service owns the boundary; `postgres.InTx` manages the lifecycle.**

- `DBTX` exposes only `Exec`, `Query`, and `QueryRow`; it does not expose `Begin`.
- `InTx` accepts a `*pgxpool.Pool`, begins a transaction, and passes `pgx.Tx` to the callback.
- A nil callback error triggers commit. A callback error is returned and triggers deferred rollback.
- Deferred rollback also runs during panic unwinding and uses a context without request cancellation so cleanup can proceed.
- Begin and commit failures are returned with operation context.
- This helper does not support nested transactions. Pass the existing transaction to participating stores instead.
- Inside the callback, use transaction-backed stores for all participating writes. Using a pool-backed store would put those writes outside the transaction.

### Example: Registration

`AuthService.Register` checks email availability, hashes the password, and calls the user store to insert the user and required password hash together. A single `INSERT` is atomic, so registration needs no explicit transaction or pool dependency. The password field is excluded from JSON with `json:"-"`.

```mermaid
sequenceDiagram
    participant H as Auth handler
    participant S as Auth service
    participant U as User store
    participant DB as PostgreSQL

    H->>S: Register(ctx, payload)
    S->>U: FindByEmail (pool)
    U->>DB: SELECT user
    DB-->>U: Result
    U-->>S: User or error
    Note over S: Reject existing email or unexpected error; otherwise hash password
    S->>U: Create(name, email, passwordHash)
    U->>DB: INSERT user with password hash
    DB-->>U: User or error
    U-->>S: User or error
    S-->>H: User or error
```

A rejected password write must leave no user behind. The email pre-check is only an early response: the database unique constraint remains authoritative when registrations race.

## 4. Validation and Errors

| Layer | Validates | Failure handling |
| --- | --- | --- |
| **Handler** | JSON shape, required fields, lengths, basic formats | Return `400 Bad Request` |
| **Service** | Business rules, permissions, application state | Translate expected domain failures into application errors |
| **Store / database** | Unique keys, foreign keys, not-null and other integrity constraints | Translate known violations; wrap unexpected failures |

Request validation uses `validation.FieldError` (`field`, `code`, `message`) and `validation.Errors`, an ordered collection that also implements `error`. Modules use `go-playground/validator/v10` built-in tags and translate its failures into these types; no custom validation rules are registered. Email follows the library's `email` rule. Password validation uses built-in `Var` checks on a trimmed copy for `required` and on `[]byte(password)` for `max=72`, so whitespace-only passwords are rejected and bcrypt's byte limit is enforced without modifying the password. Modules own normalization and error messages; platform validation types contain no HTTP details. DTO validation reports all invalid fields in declaration order, with only the first failed rule per field. Account-store defensive validation preserves these errors with `%w`; callers can recover them with `errors.As` rather than receiving a misleading account-type error.

`httpx.DecodeJSON` accepts exactly one JSON object with no unknown fields and enforces a 1 MiB limit over the complete body, including trailing data. Handlers return `httpx.DecodeError` for safe, actionable decode responses and `httpx.ValidationError` for field validation; the adapter writes the response. Invalid bodies and fields return `400`; oversized bodies return `413 PAYLOAD_TOO_LARGE`. The existing `error.code` and `error.message` envelope gains optional `details` for field-specific failures. Body-wide failures omit `details`. Do not echo submitted values or raw decoder errors, and do not make clients parse message text. See [OpenAPI](../api/openapi.yaml) for the complete contract and stable field codes.

Errors travel upward through normal Go returns, not exceptions:

1. The store translates recognized database failures into domain errors.
2. Domain sentinels stay beside their domains. Services translate expected failures into `apperror.Error` (`Code`, `Message`, optional `Details`) with a private cause accessible through `Unwrap`. Unexpected errors propagate, optionally wrapped with `%w`; `errors.Is` and `errors.As` continue to work.
3. Handlers return errors to `httpx.Adapt`; standard middleware can call `httpx.HandleError`. The transport maps known public codes to statuses, such as `409 Conflict` for `EMAIL_TAKEN`. Unknown codes, internal errors, and non-serializable details are logged and become `500 INTERNAL_ERROR` with message `internal error`; underlying causes never enter JSON.

Top-level public codes live in `apperror`; existing string values remain stable. Field-validation codes remain in `validation`. Application error messages and details must be client-safe, and details must be JSON-serializable. Validation continues to use ordered field-error arrays; other application errors may supply other JSON values. HTTP status codes belong only to the transport. A handler must not return an error after committing headers or a body; the adapter cannot replace an already-written response.

Data flows upward as **database rows → store models → service results → handler JSON**. Keep sensitive persistence fields out of public responses.

## 5. Verification

Follow the [testing guidelines](testing_guideline.md) for coverage responsibilities, case limits, naming, assertions, isolation, and review rules.

- **Unit tests** exercise dependency-free validation/parsing and distinct middleware contracts using standard `testing` and `httptest`. Small handwritten stubs are reserved for difficult-to-trigger failures; middleware lookup callbacks may verify forwarding/context and downstream execution.
- **Integration tests** use `//go:build integration`, real services/stores, and PostgreSQL for ordinary auth/session outcomes, hashing, persistence, expiry, rollback, and HTTP response/cookie contracts. Tests through `app.New` prove representative wiring and a complete lifecycle, without repeating lower-level matrices or storage assertions. Each behavioral contract has one primary owner; domain errors and their HTTP mapping are distinct contracts. Do not change production interfaces solely for testing.
- **Database isolation** uses `internal/testutil.Postgres`: explicit `.env.test`, real Goose migrations, one private schema per test, schema-only `search_path`, and schema-scoped cleanup. Setup failures are fatal, not skipped.

See [tests and checks](../README.md#tests-and-checks) for local setup and commands.
