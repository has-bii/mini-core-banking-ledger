# HTTP Conventions

HTTP is the transport boundary. These conventions apply to new and changed handlers. [OpenAPI](../api/openapi.yaml) defines the public contract; update it and the relevant tests when that contract changes.

## Responsibilities

| Layer   | Owns                                                                                                                        |
| ------- | --------------------------------------------------------------------------------------------------------------------------- |
| HTTP    | Register routes, decode requests, validate input DTOs, pass verified identity, and map results and errors to HTTP responses |
| DTO     | Normalize structured input and validate field rules that do not require application or storage state                        |
| Service | Enforce business policy and permissions, and coordinate workflows                                                           |
| Store   | Check persistence identifiers, scope queries by owner, and enforce persistence constraints                                  |

Keep handlers free of business decisions, storage calls, and repeated validation. See [DTO Conventions](dto_conventions.md), [Service Conventions](service_conventions.md), and [Store Conventions](store_conventions.md).

## Files and routing

- Start with `http.go` and `http_test.go` in the module. Use a concrete `Handler`, a `NewHandler` constructor, and private methods for operations.
- Define a small, unexported interface beside the handler for the service operations it calls. Inject dependencies through the constructor and wire implementations in the app. Don't construct services or stores in handlers.
- Register method-qualified `http.ServeMux` patterns in `Handler.Register`. Keep patterns relative to the API mount. Read path variables with `r.PathValue`.
- Wrap protected routes with `middleware.AuthGuard`. In a guarded handler, read the verified identity from request context and pass it explicitly to the service. Never take the caller's identity from the body, a path parameter, or an unverified cookie.
- Middleware can answer a request before the handler runs. The router can also respond directly to an unmatched path or unsupported method. Don't assume those responses use the handler's envelope; test them at the layer that writes them.

## Requests

1. Decode JSON bodies with `httpx.DecodeJSON`. It limits body size and requires one JSON object without unknown fields.
2. Call the DTO's `Validate` once before invoking the service. Keep scalar inputs as scalars when no DTO is needed, and keep server-owned fields out of input DTOs.
3. Pass `r.Context()` through to the service. Pass verified identity and other request arguments separately; don't replace the request context or make the service read HTTP state.

## Responses

- Use `httpx.SendSuccess` for JSON successes (`data` and optional `message`). Choose the status for the operation: `200` for a result, `201` for creation, or `204` for a successful operation with no body. Write `204` without a JSON envelope.
- Return a model directly only if its JSON shape is intentionally safe as the public contract. Otherwise map it to a response DTO in the handler. When the contract calls for an array, return `[]` for an empty collection, not `null`.
- Set headers and cookies before committing the response. Use `httpx.NoStoreHeader` for responses that must not be cached, including errors from those routes. When issuing or clearing a session cookie, keep its name, path, and domain consistent; keep cookie handling in HTTP and session lifecycle decisions in the service.
- Use `httpx.SendRequestError` for handler failures. It checks typed transport and validation errors with `errors.As`, then uses the module's mapper for known domain outcomes. Match domain sentinels with `errors.Is`, not error text. Map each recognized outcome to a status and safe detail; return an unmapped result for unexpected failures.
- Write handler errors as `application/problem+json`. Keep validation details keyed by JSON field name. Unexpected failures return a generic `500`; log the cause at the HTTP boundary without exposing it to the client. Never log passwords, tokens, or request bodies. If writing a response fails, log the failure rather than attempting a second response.

## Testing

**Handler tests are unit tests only.** Use standard `testing`, `httptest`, and small handwritten service stubs. Prefer an external test package and exercise registered routes, including their guards, through `ServeHTTP`.

- Assert observable status codes, headers, cookies, and decoded response fields. Check that invalid or unauthenticated requests do not reach the service.
- Verify forwarding of request context, path values, normalized input, and verified identity where applicable. Cover known domain errors and ensure unexpected error details stay out of the response.
- Keep the full field-rule matrix in DTO tests and business decisions in service tests. Test real persistence in store integration tests and cross-layer wiring in app integration tests. See the [README test instructions](../README.md#tests-and-checks).

Run unit tests from `backend/`:

```sh
go test -race ./...
```
