# HTTP Conventions

## Roles

- **Handler.** Decode and validate requests, then pass `r.Context()`, verified identity, and request arguments explicitly to the [service](service_conventions.md#roles) for business decisions and persistence. For protected routes, read the verified user ID from context, never from caller-supplied fields or an unverified cookie.
- **Response writer.** The handler, middleware, or router that writes the response owns its headers and error body. Middleware and the router can respond before the handler runs.

## How to write

1. **Wire the handler.** In the module's `http.go`, define a concrete `Handler`, a `NewHandler` constructor, and private operation methods. Define a small, unexported interface for the service methods the handler calls. Inject the service and wire the concrete implementation in `internal/app/app.go`.
2. **Register the route.** Add a method-qualified `http.ServeMux` pattern in `Handler.Register`, relative to the API mount (for example, `GET /wallets/{id}`). Read path variables with `r.PathValue`. Wrap protected routes with `middleware.AuthGuard`.
3. **Decode and call.** Use `httpx.DecodeJSON` for JSON input: it limits body size and requires one UTF-8 JSON object with case-sensitive field names, no unknown fields, and no duplicate member names. It rejects unpaired UTF-16 surrogate escapes in JSON strings. Decode into a pointer to the input DTO, call its `Validate` once, then pass small, read-only DTOs to the service by value. Follow [DTO conventions](dto_conventions.md#roles) for which values belong in the input DTO.
4. **Write success.** Use `httpx.SendSuccess` for JSON (`data` and optional `message`): `200` for a result, `201` for creation. Write `204` without a body. Return a model only if its JSON shape is safe and intentional as the public contract; otherwise map it to a [response DTO](dto_conventions.md#how-to-write). Return `[]`, not `null`, when the contract calls for an array.
5. **Set headers and cookies before committing the response.** Set `httpx.NoStoreHeader` before any handler response on a non-cacheable route, including errors the handler writes. Keep the session cookie's name, path, and domain consistent when issuing and clearing it.
6. **Update the contract.** When the public route, input, or response changes, update [OpenAPI](../api/openapi.yaml) and the affected tests.

## Errors

- **Transport and validation.** Use `httpx.SendRequestError` to send `application/problem+json`. It checks typed transport and validation errors with `errors.As` before the module's mapper. Keep validation details keyed by JSON field name.
- **Domain errors.** Match sentinels with `errors.Is` and map them to a status and safe detail. Leave unexpected failures unmapped for a generic `500`; log their causes without passwords, tokens, or request bodies.
- **Response encoding failure.** The JSON helpers log the failure and send a generic `500` before committing headers. This fallback includes the request path in `instance` and drops any staged `Set-Cookie` header.
- **Panic recovery.** Before the response is committed, middleware drops any staged `Set-Cookie` header and sends a generic `500`. After commitment, it aborts the response instead of writing a second one.
- **Write failure.** The JSON helpers log the failure without attempting a second response.

## How to test

1. **Set up `http_test.go`.** Inject a handwritten service stub and register the real routes and guards on a mux. Use [unit test conventions](unit_test_conventions.md#http-route) for the `httptest` setup and example.
2. **Serve a request.** Call `ServeHTTP` with an `httptest` request and recorder. On a valid request, assert forwarded context, path values, normalized input, and verified identity where relevant. For invalid or unauthenticated requests, assert that the service wasn't called.
3. **Check the wire contract.** Assert status, relevant headers and cookies, and decoded response fields. Exercise recognized domain errors and verify that unexpected details stay out of the response. Test middleware- or router-written responses at the layer that writes them; keep the full field-rule matrix out of handler tests.

Run handler unit tests from `backend/`:

```sh
go test -race ./internal/modules/wallet -run '^TestHandler_'
```
