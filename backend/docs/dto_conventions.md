# DTO Conventions

## Roles

- **Input.** Group structured caller-settable fields; keep scalar arguments when clear. Limit normalization and validation to rules that need no application or storage state. The handler validates the DTO and passes verified identity and server-owned values separately. Reuse the validated DTO when downstream calls need the same fields. Pass small, read-only DTOs by value after validation.
- **Output.** A response DTO defines the public shape separately from the model. Map to it in the handler, inline or with a private helper such as `toWalletResponse`.

## How to write

1. **Define the shape.** Put module DTOs in `dto.go`; split into `dto_<concern>.go` only for distinct concerns. Name input types for the operation (`CreateWalletDTO`) and give each field an explicit JSON tag (`wallet_type`).
2. **Normalize and validate.** Expose `Validate() error`. Use a pointer receiver for in-place normalization and document what changes, including on error. Normalize only fields whose meaning allows it, then validate the normalized values. Preserve credentials and opaque tokens as supplied.
3. **Define an output shape when needed.** Use a named response DTO such as `walletResponse` when fields must be hidden, renamed, or reshaped, or when the wire contract needs to evolve separately. Treat changes to JSON field names, types, or omission behavior as contract changes.

## Errors

- **Invalid input.** Configure the module validator with `UseJSONName`; use `FormatErrors` and `NewErrorValidation` to return `*validator.ErrorValidation` with JSON field names and safe messages.
- **Validator misuse.** Treat misuse (such as `InvalidValidationError`) as an internal error; wrap with `%w` when adding context.
- **Handoff.** The [response writer](http_conventions.md#errors) owns statuses and logging. Keep submitted secrets out of DTO errors.

## How to test

1. **Set up `dto_test.go`.** Create a fresh DTO for each case. Use [unit test conventions](unit_test_conventions.md#dto-validation) for a named subtest example.
2. **Call `Validate` directly.** Cover valid input, missing fields, and rule boundaries. Assert the normalized field values, including on error when documented, and verify that inputs meant to be preserved stay unchanged.
3. **Check field errors.** Use `errors.As` to extract `*validator.ErrorValidation`. Assert JSON field keys and client-facing details, rather than internal validator messages.

Run DTO unit tests from `backend/`:

```sh
go test -race ./internal/modules/wallet -run '^TestCreateWalletDTO_'
```
