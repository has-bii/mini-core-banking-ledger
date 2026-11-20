# DTO Conventions

A data transfer object (DTO) groups fields for an operation or defines a wire shape. These conventions apply to new and changed DTOs. Use [wallet](../internal/modules/wallet/dto.go) as the reference; existing DTOs do not need to move or be renamed just to match this guide.

## Responsibilities

| Layer | Owns |
| --- | --- |
| HTTP | Decode the request, call `Validate` once before invoking the service, pass verified identity separately, and map results and errors to HTTP responses |
| Input DTO | Fields the caller can set, wire tags, field normalization, and validation that does not need application or storage state |
| Service | Business policy, permissions, and workflow decisions |
| Store | Identifier checks, ownership predicates, and persistence constraints |

Use input DTOs for structured data. Keep scalar arguments when they read clearly; wallet's GET takes a path ID and needs no DTO. A module can pass the same validated DTO through its service and store when both need those fields. Don't create a second type just to copy them. Wallet creation takes `Name` and `WalletType` from `CreateWalletDTO` and takes `userID` separately from the verified session. Server-owned values such as balance and sort order are not request fields. See [Service Conventions](service_conventions.md) and [Store Conventions](store_conventions.md).

For output, return the model directly only if its JSON shape is intentionally safe and suitable as the public contract. Introduce a response DTO when fields need to be hidden, renamed, or reshaped, or when the wire contract needs to evolve independently. Do that mapping in HTTP; services and stores still return domain values or workflow results. Wallet intentionally returns its existing [`Wallet` JSON shape](../internal/modules/wallet/wallet.go), not `CreateWalletDTO`.

## Files and names

- Start with `dto.go` and `dto_test.go` in the module. Split by a distinct concern only when it improves navigation: `dto_<concern>.go` and `dto_<concern>_test.go`. Keep module-specific DTOs out of shared `dto` packages.
- Name input types for the operation, such as `CreateWalletDTO`. Use explicit JSON tags (for example, `wallet_type`); review changes to field names, types, or omission behavior as changes to the wire contract.
- Use `Validate() error` for input validation. Use a pointer receiver when normalizing in place, and document which fields change and whether they remain changed on error.
- Keep supporting functions private and name them for what they do. Use a named response type such as `walletResponse` and, if mapping needs a helper, a name such as `toWalletResponse`. Inline a mapping that is clearer at its call site.

## Normalization and validation

1. For JSON requests, decode into the input DTO with [`httpx.DecodeJSON`](../internal/platform/httpx/request.go). It enforces a single JSON object, rejects unknown fields, and limits the body size.
2. Call `Validate` before passing the DTO on. Normalize only fields whose expected meaning allows it, then check the normalized values. Wallet trims both fields before checking name length and characters and the case-sensitive wallet type. Don't apply blanket trimming or case conversion to credentials or opaque tokens.
3. Pass only normalized, validated input onward. Non-HTTP callers must meet the same preconditions; document non-obvious requirements on the receiving method, including non-nil pointer inputs. Services and stores neither revalidate nor silently mutate the DTO. State-dependent rules such as whether a caller may perform an operation belong in the service; uniqueness and other integrity constraints still belong in persistence.

## Errors

- Configure the module's validator with the platform's `UseJSONName` so field errors use wire names such as `wallet_type`, not `WalletType`. Use `FormatErrors` and `NewErrorValidation` to return `*validator.ErrorValidation` for field-rule failures. Don't flatten structured errors into strings or include submitted secrets in messages.
- Treat validator misuse (such as `InvalidValidationError`) as an internal failure, not a field error. Add operation context with `%w` when needed so callers can still inspect the cause.
- The [HTTP handler](../internal/modules/wallet/http.go) owns status and response formatting. It distinguishes decode errors (`httpx.HTTPError`), field errors (`400` with field details), domain errors (for example, `404` for `ErrNotFound`), and unexpected failures (generic `500`). Use `errors.As` for typed errors and `errors.Is` for domain sentinels. Keep HTTP statuses and logging out of DTO methods.

## Testing

**DTO tests are unit tests.** Use standard `testing`, preferably an external test package, with names such as `TestCreateWalletDTO_Validate` and named subtests for related cases. Keep each case independent; use `t.Parallel()` only when there is no shared mutable state.

- In [`dto_test.go`](../internal/modules/wallet/dto_test.go), assert normalized values, successful validation, relevant boundaries, and structured field errors with `errors.As`. Assert field keys and any client-facing details the contract requires, not internal validator error strings.
- In [`http_test.go`](../internal/modules/wallet/http_test.go), cover decoding failures, field-error mapping, and a successful request that forwards normalized input and the verified identity. Don't repeat the entire field-rule matrix there.
- Test business decisions in service tests and persistence behavior in store integration tests. Neither repeats the DTO validation matrix.

Run the wallet reference tests from `backend/`:

```sh
go test -race ./internal/modules/wallet
```
