# Service Conventions

A service is the module's application boundary. It exposes operations callers need, enforces business policy, and coordinates dependencies. These conventions apply to new and changed services. Use [wallet](../internal/modules/wallet/service.go) for delegation and [auth](../internal/modules/auth/service.go) and [session](../internal/modules/session/service.go) for workflows.

## Responsibilities

| Layer | Owns |
| --- | --- |
| HTTP | Decoding, DTO normalization and validation, obtaining the verified caller identity, and mapping results and errors to HTTP responses |
| Service | Business eligibility and permissions, workflow sequencing, and application logic such as password hashing or token generation |
| Store | Storage lookups, owner-scoped queries, persistence error translation, and atomic persistence operations |

A service can delegate directly when the store already implements the required operation, as wallet does. Don't add a service just to give every module the same layout; `user` has a store but no service of its own.

Validate structured input once at entry. Non-HTTP callers must meet the same input preconditions. Document non-obvious requirements on the service method itself: `CreateUserWallet` requires a non-nil, normalized, validated DTO. Don't repeat DTO validation or silently mutate the input. Business policy still belongs in the service; identifier checks and owner scoping still belong in the store. Pass a verified `userID` explicitly for owner-scoped operations rather than deriving it from a request body or looking it up inside the service's context.

Keep HTTP and persistence mechanics out of services: no request or response types, SQL, Redis commands, or row scanning. Wallet creation delegates its atomic write to the store. Don't assume multiple store calls are atomic; if a workflow needs atomicity across stores, design an explicit boundary for that workflow. See [Store Conventions](store_conventions.md) for ownership predicates and transaction rules.

## Files and dependencies

- Start with `service.go` and `service_test.go` beside the rest of the module. Split only when a distinct concern makes the service easier to navigate; use `service_<concern>.go` and `service_<concern>_test.go`.
- Name the concrete type `Service` and the constructor `NewService`, returning `*Service`. The package gives the type its context: `wallet.Service`, not `wallet.WalletService`.
- Define small, unexported interfaces where they are consumed. Wallet's `walletStore` contains only the methods wallet calls. A dependency can also be another service: auth defines a `sessionService` interface for creating and revoking sessions. Don't export a service interface or add a wrapper just to support mocks.
- Inject dependencies through the constructor and wire concrete implementations in [`internal/app/app.go`](../internal/app/app.go). Services don't open connections or construct their own stores.

## Method names and signatures

Name methods for the application operation, not for the query behind it. Use `Get` or `List` where it reads naturally; make single versus collection clear without imposing a strict verb rule. The store follows its own naming conventions.

| Service operation | Persistence operation |
| --- | --- |
| `GetUserWallets` | `ListByUserID` |
| `GetUserWallet` | `GetByID` |
| `CreateUserWallet` | `Create` |
| `session.Service.Revoke` | `DeleteByTokenHash` |

Use workflow names such as `Register` or `Login` when they say more than a CRUD verb. Keep implementation helpers unexported and specific to their purpose.

- Put `ctx context.Context` first and pass it through to dependencies. Don't store a context on `Service` or replace the caller's context with `context.Background()`.
- Name parameters in dependency interfaces when their meaning or order matters. Use Go initialisms (`walletID`, `userID`) and keep argument order aligned across layers: `GetByID(ctx context.Context, walletID, userID string)`.
- Use a DTO for structured inputs and scalars where they're clear. Return domain values, collections, or a workflow result and an `error`, not HTTP responses or database rows. A command with no result can return just `error`.

## Errors

- **Pass through** errors when they already describe the outcome. `GetUserWallet` preserves `wallet.ErrNotFound` for a missing wallet or one owned by someone else; the service doesn't need an extra lookup to distinguish them.
- **Wrap** with `%w` when adding context that identifies a failing workflow step. Don't wrap merely to repeat the dependency's operation or flatten structured errors into strings.
- **Translate** when the workflow changes the caller-facing meaning. For example, `auth.Service.Me` uses `auth.ErrUnauthenticated` for a missing user but doesn't treat a malformed user ID as an authentication result. Don't turn an unexpected dependency failure into a routine business outcome.
- Keep stable domain sentinels in the module's `<module>.go`. Use `errors.Is` for sentinels and `errors.As` for typed errors, not message matching. Preserve relevant causes through wrapping or joining. Keep secrets out of error messages, and leave HTTP status mapping and request-error logging to the transport boundary.

## Testing

**Service tests are unit tests only.** Use standard `testing`, small hand-written dependency stubs, and preferably an external test package such as `wallet_test`. Use `TestService_<MethodOrBehavior>` and named subtests when a table helps. Give each case independent stub state, use `t.Context()`, and use `t.Parallel()` only when the cases don't share mutable state.

- For delegation, check what callers rely on: forwarded context and arguments (especially the owner ID), returned values, and preserved errors. Don't repeat the DTO validation matrix or test SQL filters with stubs.
- For workflows, cover decisions, early exits, error translation, and observable side effects. Check that a dependency wasn't called when a rule must stop the operation, but avoid pinning call order unless it matters to the outcome.
- Assert values and error contracts using `errors.Is` or `errors.As`. Don't assert exact internal error strings or add generated mocks.
- Test validation in DTO tests, response mapping in HTTP tests, real persistence in store integration tests, and cross-layer wiring in app integration tests. No service integration suite or service-specific integration build tag is needed.

Run the reference module's unit tests from `backend/`:

```sh
go test -race ./internal/modules/wallet
```
