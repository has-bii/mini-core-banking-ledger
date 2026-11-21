package httpx_test

import (
	"backend/internal/platform/httpx"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/iotest"
)

func TestDecodeJSON(t *testing.T) {
	t.Parallel()
	const valid = `{"email":"alice@example.com"}`
	for _, tt := range []struct {
		name   string
		body   string
		status int
		reason string
	}{
		{name: "valid object", body: valid},
		{name: "surrounding whitespace", body: " \t\n" + valid + "\r "},
		{name: "exactly at limit", body: valid + strings.Repeat(" ", (1<<20)-len(valid))},
		{name: "empty", status: http.StatusBadRequest},
		{name: "whitespace", body: " \t\r\n", status: http.StatusBadRequest},
		{name: "non-object", body: " \tnull", status: http.StatusBadRequest, reason: "body is not a JSON object (at byte offset 2)"},
		{name: "truncated JSON", body: "{", status: http.StatusBadRequest, reason: "malformed JSON (at byte offset 1)"},
		{name: "syntax error", body: " \t" + `{"email":!}`, status: http.StatusBadRequest, reason: "malformed JSON (at byte offset 11)"},
		{name: "unknown field", body: `{"client-controlled-field":"secret-value"}`, status: http.StatusBadRequest, reason: "unknown field (at byte offset 1)"},
		{name: "wrong field type", body: `{"email":123}`, status: http.StatusBadRequest, reason: "type mismatch (at byte offset 9)"},
		{name: "wrong field case", body: `{"Email":"alice@example.com"}`, status: http.StatusBadRequest},
		{name: "duplicate field", body: `{"email":"alice@example.com","email":"bob@example.com"}`, status: http.StatusBadRequest},
		{name: "invalid UTF-8", body: "{\"email\":\"\xff\"}", status: http.StatusBadRequest},
		{name: "unpaired surrogate escape", body: `{"email":"\uD800"}`, status: http.StatusBadRequest},
		{name: "multiple objects", body: " {} \t {}", status: http.StatusBadRequest, reason: "trailing data after JSON object (at byte offset 6)"},
		{name: "trailing garbage", body: valid + " x", status: http.StatusBadRequest},
		{name: "oversized value", body: `{"email":"` + strings.Repeat("a", 1<<20) + `"}`, status: http.StatusRequestEntityTooLarge},
		{name: "oversized trailing whitespace", body: valid + strings.Repeat(" ", (1<<20)-len(valid)+1), status: http.StatusRequestEntityTooLarge},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var payload struct {
				Email string `json:"email"`
			}
			err := httpx.DecodeJSON(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body)), &payload)
			if tt.status == 0 {
				if err != nil {
					t.Fatalf("DecodeJSON() = %v, want success", err)
				}
				if payload.Email != "alice@example.com" {
					t.Errorf("email = %q, want alice@example.com", payload.Email)
				}
				return
			}
			var httpErr *httpx.HTTPError
			if !errors.As(err, &httpErr) {
				t.Fatalf("DecodeJSON() = %v, want HTTPError", err)
			}
			if httpErr.Status != tt.status {
				t.Errorf("status = %d, want %d", httpErr.Status, tt.status)
			}
			if httpErr.Message != "Invalid request body" || err.Error() != httpErr.Message {
				t.Errorf("error = %q, public detail = %q, want Invalid request body", err, httpErr.Message)
			}
			cause := errors.Unwrap(err)
			if cause == nil {
				t.Fatal("DecodeJSON() error has no diagnostic cause")
			}
			if tt.reason != "" && cause.Error() != tt.reason {
				t.Errorf("cause = %q, want %q", cause, tt.reason)
			}
		})
	}
}

func TestDecodeJSON_PreservesUnexpectedErrors(t *testing.T) {
	t.Parallel()
	t.Run("body read failure", func(t *testing.T) {
		t.Parallel()
		cause := errors.New("read failed")
		r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", iotest.ErrReader(cause))
		var payload struct{}
		if err := httpx.DecodeJSON(httptest.NewRecorder(), r, &payload); err != cause {
			t.Errorf("DecodeJSON() = %v, want original read error", err)
		}
	})
	t.Run("invalid destination", func(t *testing.T) {
		t.Parallel()
		r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/", strings.NewReader(`{}`))
		err := httpx.DecodeJSON(httptest.NewRecorder(), r, (*struct{})(nil))
		var httpErr *httpx.HTTPError
		if err == nil || errors.As(err, &httpErr) {
			t.Errorf("DecodeJSON() = %v, want an unexpected error, not an HTTPError", err)
		}
	})
}

func TestDecodeJSON_RequiresObjectForAnyDestination(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		body string
	}{
		{name: "null", body: "null"},
		{name: "array", body: "[]"},
		{name: "string", body: `"text"`},
		{name: "number", body: "123"},
		{name: "boolean", body: "true"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var payload any
			err := httpx.DecodeJSON(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body)), &payload)
			var httpErr *httpx.HTTPError
			if !errors.As(err, &httpErr) {
				t.Fatalf("DecodeJSON() = %v, want HTTPError", err)
			}
			if httpErr.Status != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", httpErr.Status)
			}
		})
	}
}
