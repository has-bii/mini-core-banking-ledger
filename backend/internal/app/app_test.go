package app

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRegistrationRoute(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := New("dev", nil, log)
	req := httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("registration: got %d, %s", rec.Code, rec.Body.String())
	}
	req = httptest.NewRequest(http.MethodPost, "https://bank.example/auth/register", strings.NewReader(`{}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "https://evil.example")
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("cross-origin request got %d; want 403", rec.Code)
	}
}
