# Unit Test Conventions

## Roles

Exercise a layer's public entry point without a database, Redis, an environment file, or a listening HTTP server. Use [integration tests](integration_test_conventions.md#roles) for real persistence or assembled app wiring.

## How to write

1. **Place the test.** Match the source filename (`http_test.go`, `dto_test.go`, `service_test.go`, or the matching concern suffix). Prefer an external package such as `wallet_test`. Use standard `testing` without a build tag.
2. **Isolate the case.** Use small handwritten stubs for the interfaces the code consumes. Give each case fresh input and stub state. Use named `t.Run` subtests for related cases, `t.Context()` for calls that take context, and `t.Parallel()` only when cases have no shared mutable state.
3. **Assert behavior.** Check the result and any dependency calls that are part of the contract. Pass the current `*testing.T` into helpers and call `t.Helper()` there.

## Errors

- **HTTP rejection and public error responses:** follow the [handler checks](http_conventions.md#how-to-test).
- **Structured field errors and normalization on failure:** follow the [DTO checks](dto_conventions.md#how-to-test).
- **Error preservation, translation, and early exits:** follow the [service checks](service_conventions.md#how-to-test).

## HTTP route

Seed a verified identity when testing a protected route. Use a fake session lookup instead when session resolution is part of the case. This excerpt uses the existing `walletServiceStub` in `wallet/http_test.go`; the stub records the context if called.

```go
svc := &walletServiceStub{}
mux := http.NewServeMux()
wallet.NewHandler(svc, slog.Default()).Register(mux)
ctx := middleware.WithUserID(t.Context(), walletTestUserID)
req := httptest.NewRequestWithContext(ctx, http.MethodPost, "/wallets", strings.NewReader(`{}`))
rec := httptest.NewRecorder()

mux.ServeHTTP(rec, req)

if rec.Code != http.StatusBadRequest || rec.Header().Get("Content-Type") != "application/problem+json" {
	t.Fatalf("status = %d, headers = %v", rec.Code, rec.Header())
}
if svc.ctx != nil {
	t.Error("invalid payload reached the service")
}
```

## DTO validation

Build the value inside the subtest. This example checks a structured field error and the normalized value after the same call.

```go
t.Run("missing wallet type", func(t *testing.T) {
	input := wallet.CreateWalletDTO{Name: " Travel "}
	err := input.Validate()
	var invalid *validator.ErrorValidation
	if !errors.As(err, &invalid) || invalid.Errors == nil ||
		!reflect.DeepEqual((*invalid.Errors)["wallet_type"], []string{"This field is required"}) {
		t.Fatalf("Validate() = %v, want required error for wallet_type", err)
	}
	if input.Name != "Travel" {
		t.Errorf("name = %q, want Travel", input.Name)
	}
})
```

## Service delegation

This example uses the existing `walletStoreStub` in `wallet/service_test.go` to check error preservation and owner forwarding together.

```go
ctx := t.Context()
const ownerID = "owner"
cause := fmt.Errorf("load wallet: %w", wallet.ErrNotFound)
store := &walletStoreStub{err: cause}

_, err := wallet.NewService(store).GetUserWallet(ctx, "wallet", ownerID)

if !errors.Is(err, cause) || !errors.Is(err, wallet.ErrNotFound) {
	t.Fatalf("GetUserWallet() error = %v, want preserved cause and ErrNotFound", err)
}
if store.ctx != ctx || store.id != "wallet" || store.userID != ownerID {
	t.Error("store did not receive the original context and IDs")
}
```

Run the unit suite from `backend/`:

```sh
go test -race ./...
```
