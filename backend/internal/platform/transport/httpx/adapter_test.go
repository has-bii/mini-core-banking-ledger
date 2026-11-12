package httpx_test

import (
	"backend/internal/platform/apperror"
	"backend/internal/platform/transport/httpx"
	"bytes"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestAdapt_StatusMapping(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		code   string
		status int
	}{
		{"INVALID_JSON", 400},
		{"VALIDATION_ERROR", 400},
		{"PAYLOAD_TOO_LARGE", 413},
		{"UNAUTHENTICATED", 401},
		{"INVALID_CREDENTIALS", 401},
		{"FORBIDDEN", 403},
		{"EMAIL_TAKEN", 409},
		{"TOO_MANY_REQUESTS", 429},
		{"UNPROCESSABLE_ENTITY", 422},
	} {
		t.Run(tt.code, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			httpx.Adapt(func(http.ResponseWriter, *http.Request) error {
				return apperror.New(tt.code, "safe message", nil, nil)
			}, nil).ServeHTTP(w, httptest.NewRequest("GET", "/", nil))
			want := fmt.Sprintf(`{"error":{"code":%q,"message":"safe message"}}`, tt.code)
			if w.Code != tt.status || strings.TrimSpace(w.Body.String()) != want {
				t.Errorf("response = %d %s; want %d %s", w.Code, w.Body.String(), tt.status, want)
			}
		})
	}
}

func TestAdapt(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name   string
		err    error
		status int
		body   string
		logged bool
	}{
		{name: "success", status: 204},
		{name: "wrapped application error", err: fmt.Errorf("lookup: %w", apperror.New("NOT_FOUND", "account not found", map[string]string{"resource": "account"}, errors.New("private cause"))), status: 404, body: `{"error":{"code":"NOT_FOUND","message":"account not found","details":{"resource":"account"}}}`},
		{name: "unexpected failure", err: errors.New("private cause"), status: 500, body: `{"error":{"code":"INTERNAL_ERROR","message":"internal error"}}`, logged: true},
		{name: "unknown code", err: apperror.New("UNKNOWN", "private cause", "secret", nil), status: 500, body: `{"error":{"code":"INTERNAL_ERROR","message":"internal error"}}`, logged: true},
		{name: "internal application error", err: apperror.New("INTERNAL_ERROR", "operation failed", "secret", errors.New("private cause")), status: 500, body: `{"error":{"code":"INTERNAL_ERROR","message":"internal error"}}`, logged: true},
		{name: "unencodable details", err: apperror.New("NOT_FOUND", "missing", make(chan int), errors.New("private cause")), status: 500, body: `{"error":{"code":"INTERNAL_ERROR","message":"internal error"}}`, logged: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var logs bytes.Buffer
			logger := slog.New(slog.NewTextHandler(&logs, nil))
			r := httptest.NewRequest("GET", "/accounts", nil).WithContext(t.Context())
			w := httptest.NewRecorder()
			httpx.Adapt(func(w http.ResponseWriter, got *http.Request) error {
				if got != r {
					t.Error("adapter replaced request")
				}
				if tt.err == nil {
					w.WriteHeader(http.StatusNoContent)
				}
				return tt.err
			}, logger).ServeHTTP(w, r)
			if w.Code != tt.status || strings.TrimSpace(w.Body.String()) != tt.body {
				t.Errorf("response = %d %s; want %d %s", w.Code, w.Body.String(), tt.status, tt.body)
			}
			if tt.err != nil && w.Header().Get("Content-Type") != "application/json" {
				t.Error("missing JSON content type")
			}
			if tt.logged != (logs.Len() > 0) {
				t.Errorf("logs = %q; want logged %v", logs.String(), tt.logged)
			}
			if tt.logged && !strings.Contains(logs.String(), "private cause") {
				t.Errorf("logs lost underlying failure: %s", &logs)
			}
			if tt.logged && !strings.Contains(logs.String(), "path=/accounts") {
				t.Errorf("logs lack request context: %s", &logs)
			}
		})
	}
}
