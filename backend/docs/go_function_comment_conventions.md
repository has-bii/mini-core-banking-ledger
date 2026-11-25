# Go Function Comment Conventions

## How to write

1. **Document exported production functions and methods.** Put a `//` doc comment directly above the declaration. Start with the exact function or method name and use a complete sentence. Comment unexported helpers when their names leave important intent unclear.
2. **State the caller-facing contract.** Say what the function returns or what side effect it performs. Call out `nil` results, fallbacks, or errors when callers need to handle them. Keep implementation rationale next to the code it explains.
3. **Show a return value when its shape is useful.** Use one representative value for a structured or surprising result. Name the input or condition, then show the value as Go code after a blank `//` line. Indent the code with a tab after `//` so it renders as a code block.
4. **Skip doc comments on test functions.** Use test and subtest names to describe the behavior under test.

## Return-value example

From [`FormatErrors`](../internal/platform/validator/validator.go), with JSON field names enabled:

```go
// FormatErrors groups validation messages by field. It returns nil
// unless err is or wraps a validator.ValidationErrors value.
//
// For example, an invalid email field can produce:
//
//	ErrorField{"email": {"Must be a valid email address"}}
```

See [Go Doc Comments](https://go.dev/doc/comment#func) for more examples.
