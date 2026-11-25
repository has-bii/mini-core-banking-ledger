# Backend agent guide

## Working rules

- **Layer boundaries:** Validate input at the HTTP boundary, keep workflows in services and persistence in stores, and wire concrete dependencies in `internal/app`.
- **API contract:** When public HTTP behavior changes, update [OpenAPI](api/openapi.yaml) and the affected tests in the same change.
- **Verification:** Test at the lowest layer that proves the behavior. For Go changes, run `golangci-lint run` from the backend root. Use the [README checks](README.md#tests-and-checks) for other commands and setup. Run the relevant checks and report any you couldn't run.

## Backend structure

- [`cmd/api/`](cmd/api/) — API entry point; [`internal/app/`](internal/app/) — routes and dependency wiring.
- [`internal/modules/`](internal/modules/) — feature code, grouped by domain.
- [`internal/config/`](internal/config/) — runtime settings; [`internal/platform/`](internal/platform/) — shared infrastructure.
- [`internal/testutil/`](internal/testutil/) — shared test fixtures.
- [`api/`](api/) — OpenAPI contract; [`migrations/`](migrations/) — Goose SQL migrations.

## Read before changing

- **Go function comments** — Doc comments and return-value examples: [Go function comment conventions](docs/go_function_comment_conventions.md).
- **HTTP** — Routes, handlers, middleware, cookies, or response contracts: [HTTP conventions](docs/http_conventions.md).
- **DTO** — Request fields, normalization, validation, or response shapes: [DTO conventions](docs/dto_conventions.md).
- **Service** — Workflows, permissions, dependency coordination, or domain error mapping: [service conventions](docs/service_conventions.md).
- **Store** — SQL, Redis, ownership filters, transactions, or persistence errors: [store conventions](docs/store_conventions.md).
- **Unit tests** — Field rules, HTTP behavior, or service workflows without real backends: [unit test conventions](docs/unit_test_conventions.md).
- **Integration tests** — Migrations, real PostgreSQL or Redis behavior, or app wiring: [integration test conventions](docs/integration_test_conventions.md).
