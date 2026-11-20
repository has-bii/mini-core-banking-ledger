# HTTP Conventions

## How to write

1. **Wire the handler.** In the module's `http.go`, define a concrete `Handler`, a `NewHandler` constructor, and private operation methods. Define a small, unexported interface for the service methods the handler calls. Inject the service and wire the concrete implementation in `internal/app/app.go`.
2. **Register the route.** Add a method-qualified `http.ServeMux` pattern in `Handler.Register`, relative to the API mount (for example, `GET /wallets/{id}`). Read path variables with `r.PathValue`. Wrap protected routes with `middleware.AuthGuard`; read the verified user ID from context, never from caller-supplied request fields or an unverified cookie.
3. **Decode and call.** Use `httpx.DecodeJSON` for JSON input: it limits body size and requires one object with no unknown fields. Call the input DTO's `Validate` once. Keep scalar inputs as scalars and keep server-owned values out of the DTO. Pass `r.Context()`, verified identity, and request arguments explicitly to the service. Delegate business decisions and persistence through the service.
4. **Write success.** Use `httpx.SendSuccess` for JSON (`data` and optional `message`): `200` for a result, `201` for creation. Write `204` without a body. Return a model only if its JSON shape is safe and intentional as the public contract; otherwise map it to a response DTO. Return `[]`, not `null`, when the contract calls for an array.
5. **Set headers and cookies before committing the response.** Set `httpx.NoStoreHeader` before any handler response on a non-cacheable route, including errors the handler writes. Keep the session cookie's name, path, and domain consistent when issuing and clearing it. Middleware and the router can respond before the handler runs; their headers and error bodies belong to the writer.
6. **Write failures.** Use `httpx.SendRequestError` to send `application/problem+json`. It checks typed transport and validation errors with `errors.As` before the module's mapper. Match domain sentinels with `errors.Is` and map them to a status and safe detail; leave unexpected failures unmapped for a generic `500`. Keep validation details keyed by JSON field name. Log unexpected causes without passwords, tokens, or request bodies. If writing fails, log the failure instead of attempting a second response.
7. **Update the contract.** When the public route, input, or response changes, update [OpenAPI](../api/openapi.yaml) and the affected tests.

## How to test

1. **Set up `http_test.go`.** Inject a handwritten service stub and register the real routes and guards on a mux. Use [unit test conventions](unit_test_conventions.md#http-route) for the `httptest` setup and example.
2. **Serve a request.** Call `ServeHTTP` with an `httptest` request and recorder. On a valid request, assert forwarded context, path values, normalized input, and verified identity where relevant. For invalid or unauthenticated requests, assert that the service wasn't called.
3. **Check the wire contract.** Assert status, relevant headers and cookies, and decoded response fields. Exercise recognized domain errors and verify that unexpected details stay out of the response. Test middleware- or router-written responses at the layer that writes them; keep the full field-rule matrix out of handler tests.

Run handler unit tests from `backend/`:

```sh
go test -race ./internal/modules/wallet -run '^TestHandler_'
```
