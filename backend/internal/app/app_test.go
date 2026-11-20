package app_test

import (
	"backend/internal/app"
	"backend/internal/config"

	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew_CORSOrigin(t *testing.T) {
	t.Parallel()

	const origin = "https://app.example.com"
	handler, err := app.New(config.Config{
		Env: "test", CORSOrigin: origin, RatelimitCapacity: 50, RatelimitRefillRate: 10,
	}, nil, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, origin, wantOrigin string
		wantStatus               int
	}{
		{"configured allowed", origin, origin, http.StatusNotFound},
		{"unlisted denied", "https://untrusted.example.com", "", http.StatusForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, "/api/not-found", nil)
			request.Header.Set("Origin", tc.origin)
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)
			if response.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, tc.wantStatus)
			}
			if got := response.Header().Get("Access-Control-Allow-Origin"); got != tc.wantOrigin {
				t.Fatalf("CORS origin = %q, want %q", got, tc.wantOrigin)
			}
		})
	}
}

func TestNew_RateLimitCORS(t *testing.T) {
	t.Parallel()

	const origin = "http://localhost:3000"
	handler, err := app.New(config.Config{
		Env: "test", CORSOrigin: origin, RatelimitCapacity: 1, RatelimitRefillRate: 0.001,
	}, nil, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}

	// Unregistered routes exercise middleware without touching the database.
	const path = "/api/not-found"
	get := func() *httptest.ResponseRecorder {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		request.Header.Set("Origin", origin)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		return response
	}

	if response := get(); response.Code != http.StatusNotFound {
		t.Fatalf("first request status = %d, want 404", response.Code)
	}
	limited := get()
	if limited.Code != http.StatusTooManyRequests {
		t.Fatalf("second request status = %d, want 429", limited.Code)
	}
	if got := limited.Header().Get("Access-Control-Allow-Origin"); got != origin {
		t.Errorf("429 CORS origin = %q, want %q", got, origin)
	}
	if got := limited.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
		t.Errorf("429 allow credentials = %q, want true", got)
	}

	preflight := httptest.NewRequest(http.MethodOptions, path, nil)
	preflight.Header.Set("Origin", origin)
	preflight.Header.Set("Access-Control-Request-Method", http.MethodPost)
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, preflight)
	if response.Code != http.StatusNoContent {
		t.Errorf("preflight status = %d, want 204", response.Code)
	}
}
