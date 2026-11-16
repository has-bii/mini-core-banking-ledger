package httpx_test

import (
	"backend/internal/platform/v2/httpx"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeJSON(t *testing.T) {
	t.Parallel()
	const valid = `{"email":"alice@example.com"}`
	for _, tt := range []struct {
		name   string
		body   string
		status int
	}{
		{name: "valid object", body: valid},
		{name: "surrounding whitespace", body: " \t\n" + valid + "\r "},
		{name: "exactly at limit", body: valid + strings.Repeat(" ", (1<<20)-len(valid))},
		{name: "empty", status: http.StatusBadRequest},
		{name: "whitespace", body: " \t\r\n", status: http.StatusBadRequest},
		{name: "null struct", body: "null", status: http.StatusBadRequest},
		{name: "malformed", body: "{", status: http.StatusBadRequest},
		{name: "unknown field", body: `{"unknown":true}`, status: http.StatusBadRequest},
		{name: "wrong field type", body: `{"email":123}`, status: http.StatusBadRequest},
		{name: "multiple objects", body: valid + " {}", status: http.StatusBadRequest},
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
			var bodyErr *httpx.MalformedRequestBody
			if !errors.As(err, &bodyErr) {
				t.Fatalf("DecodeJSON() = %v, want MalformedRequestBody", err)
			}
			if bodyErr.Status != tt.status {
				t.Errorf("status = %d, want %d", bodyErr.Status, tt.status)
			}
		})
	}
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
			var bodyErr *httpx.MalformedRequestBody
			if !errors.As(err, &bodyErr) {
				t.Fatalf("DecodeJSON() = %v, want MalformedRequestBody", err)
			}
			if bodyErr.Status != http.StatusBadRequest {
				t.Errorf("status = %d, want 400", bodyErr.Status)
			}
		})
	}
}
