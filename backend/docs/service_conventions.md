# Service Conventions

## Roles

- **Input.** Take verified identity explicitly and use the caller's normalized, validated DTO as supplied.
- **Workflow.** Enforce eligibility and permissions, perform application logic such as password hashing or token generation, and coordinate dependencies. Delegate directly when one dependency already implements the operation. Leave transport concerns to [HTTP](http_conventions.md#roles) and SQL, Redis commands, and row scanning to [stores](store_conventions.md#roles).

## How to write

1. **Wire the service.** Use `service.go` for a concrete `Service` and `NewService` returning `*Service`. Split by concern into `service_<concern>.go` only when it improves navigation. Define small, unexported interfaces for the dependency methods the service calls. Inject dependencies through the constructor and wire real implementations in `internal/app/app.go`. Add a service when callers need an application operation, not just to mirror another module's layout.
2. **Name and type the operation.** Use application names such as `Register`, `Revoke`, or `GetUserWallet`. Put `ctx context.Context` first and pass it through unchanged. Name dependency parameters when order matters and keep that order consistent (for example, `walletID, userID`). Accept scalars when clear and structured input as a DTO. Document preconditions such as a non-nil, normalized, validated DTO. Return domain values, collections, workflow results, or just `error` for commands with no result.
3. **Choose the atomic boundary.** Use a single atomic persistence operation or an explicit transaction boundary when required; sequential store calls aren't atomic.

## Errors

- **Preserve or wrap.** Pass through errors that already describe the outcome. Wrap with `%w` to identify a failing workflow step. Keep causes inspectable and secrets out of messages.
- **Translate only when meaning changes.** `Me` returns `auth.ErrUnauthenticated` for a missing user, but preserves `user.ErrInvalidUserID` as an internal contract failure.
- **Match and define.** Match stable module sentinels with `errors.Is` and typed errors with `errors.As`. Define sentinels in `<module>.go`.

## How to test

1. **Set up `service_test.go`.** Call `NewService` with handwritten dependency stubs and exercise the exported method with `t.Context()`. Use [unit test conventions](unit_test_conventions.md#service-delegation) for a stub example.
2. **Check delegation.** Assert forwarded context and arguments, especially the owner ID, plus returned values and preserved errors. Inspect errors with `errors.Is` or `errors.As`, not internal message text.
3. **Check workflows.** Cover business decisions, early exits, error translation, and observable side effects. Assert that a dependency wasn't called when a rule must stop the operation; assert call order only when the behavior requires it. Keep field-rule matrices and query behavior out of service tests.

Run service unit tests from `backend/`:

```sh
go test -race ./internal/modules/wallet -run '^TestService_'
```
