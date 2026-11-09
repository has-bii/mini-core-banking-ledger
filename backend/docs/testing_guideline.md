# Backend Testing Guidelines

Keep tests small, independent, and focused on observable behavior. Test rules and contracts, not implementation details or coverage percentages.

## 1. What to test

| Boundary | Test type | Main checks |
| --- | --- | --- |
| Payload validation | Unit | Required fields, normalization, format, length boundaries |
| HTTP handler | Unit with `httptest` | JSON decoding, status/error mapping, response shape, cookie security |
| Service | Unit with handwritten interface stubs | Business decisions, domain errors, dependency failures |
| Store | Integration with PostgreSQL | Persistence, not-found results, constraint/error translation |
| Transactional workflow | Integration with PostgreSQL | Successful writes and rollback when a later write fails |
| Application wiring | Integration through `app.New` | Representative complete workflows, not every input permutation |

- Cover each rule at the lowest boundary that can prove it. Add application-level tests for wiring and complete workflows; do not repeat the full case matrix at every layer.
- Exercise real migrations, SQL, transactions, and constraints in integration tests. Do not mock SQL to prove database behavior.
- For authentication, verify password/token hashing, rejected credentials, session expiry/revocation, cookie flags, and absence of secrets in JSON or unexpected-error responses.
- For a bug fix, add a regression test that fails before the fix where practical.

## 2. Case limits

These are review guardrails, not correctness limits:

- Start with one success case, each distinct failure outcome, and relevant boundary values. A single focused test is fine; do not turn everything into a table.
- **Default maximum: 12 cases per table for one behavior.** Above 12, split by meaningful behavior or add a one-line comment explaining why the larger table is necessary. Do not split into arbitrary numbered batches to satisfy the limit.
- For a length or amount limit, prefer just below, exactly at, and just above the boundary over many arbitrary values.
- Do not multiply independent inputs into a Cartesian product unless their interaction changes behavior.
- Multiple assertions are appropriate when they establish one result, such as a session cookie's security attributes. There is no assertion-count quota.
- Never drop a security, data-integrity, money, or regression case to fit a case budget. Review large files/functions for separate behaviors rather than enforcing line-count caps.

## 3. Structure and assertions

- Put tests beside their source: `service.go` → `service_test.go`. Use `service_integration_test.go` for the integration counterpart, with `//go:build integration` as the first line.
- Prefer external test packages (`package auth_test`) and exported behavior. Do not export production internals or add interfaces solely to make tests easier.
- Name tests `Test<Type>_<Behavior>` or `Test<Function>`. Keep test order aligned with the source where practical.
- Table-driven cases must have descriptive names passed to `t.Run`; name the behavior, not `case1` or `test2`.
- Use standard `testing`, `httptest`, and small handwritten stubs for existing interfaces. No generated mocks, suite framework, or new dependency without a demonstrated need.
- Assert returned values, state visible through public interfaces, and error contracts. Use `errors.Is`/`errors.As` for wrapped errors; match exact messages only when they are part of the HTTP contract.
- Verify success and error response contents, not just status codes or whether an error exists. Use independent expected values, not the production helper being tested.
- Use `t.Fatal`/`t.Fatalf` for failed setup or prerequisites; use `t.Error`/`t.Errorf` for independent checks. Helpers accepting `*testing.T` must call `t.Helper()`.
- Prefer local setup. Extract helpers only for repeated setup or assertions; reuse `internal/testutil.Postgres` rather than building another fixture framework.

## 4. Independence and determinism

- Every test must run alone and in any order. A lifecycle test may have sequential steps internally, but must create its own starting state.
- Use `t.Parallel()` only when tests do not share mutable state. Tests that change process-wide state with `t.Setenv` must not run in parallel or have parallel ancestors.
- Bind assertions and stubs that report failures to the current subtest's `*testing.T`, not the parent test's.
- Use `t.Context()` for test operations and `t.Cleanup` for resources. Cleanup that must outlive the test context needs its own bounded context.
- Do not use sleeps for synchronization or expiry. Set past timestamps for expiry tests and use channels or other explicit synchronization for concurrent work.
- Keep real password hashing and token generation in behavior tests. Assert validity, hashing, and lifetime rather than exact random tokens or timing measurements.
- Build fixtures within each test. Do not share database records, depend on seeded IDs, or ignore setup errors.

## 5. Integration database safety

- Use a dedicated PostgreSQL 18 test database. Read `DATABASE_URL` explicitly from the ignored `backend/.env.test`; never fall back to `.env` or the shell's `DATABASE_URL`.
- Call `testutil.Postgres(t)` for an isolated pool. It creates a unique schema, applies actual migrations with the existing Goose CLI, and configures schema-only `search_path` with no `public` fallback on pool connections and migrations.
- Register cleanup immediately after acquiring resources. Close pool connections before dropping only the schema created for that test. Never truncate shared tables or drop a shared database/schema.
- Schema isolation does not make a production or development database an acceptable test target. Missing configuration, Goose, PostgreSQL, or required permissions must fail requested integration tests, not skip them.
- Database fixtures may force failures with schema-local constraints or set timestamps directly. Do not add production switches or transaction abstractions just for tests.
- A forcibly terminated test process may leave its schema behind. Verify no tests are using it before manual removal.
- The current auth/session tests need neither Redis nor a running API server. Unit tests must not load environment files or require external services.

## 6. Commands and review

Run from `backend/`; see [setup instructions](../README.md#tests-and-checks).

```sh
go test -race ./...
go test -race -count=1 -tags=integration ./... # Unit and integration tests
go build ./...
go vet ./...
go vet -tags=integration ./...
```

During development, run the affected package or named test first. Before finishing, run the full relevant suite; use `-shuffle=on` when checking order independence. Report checks not run and their reason.

Review whether each case catches a distinct regression, fixtures are isolated, cleanup is safe, and sensitive values cannot leak. Coverage reports may locate gaps, but no percentage gate or minimum test count is required. CI, benchmarks, fuzzing, snapshots, and extra test tooling are not default requirements; add them only for a concrete need.
