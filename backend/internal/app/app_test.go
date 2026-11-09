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

func TestNew_RateLimitCORS(t *testing.T) {
	t.Parallel()

	const origin = "http://localhost:3000"
	handler, err := app.New(config.Config{Env: "test", CORSOrigin: origin}, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}

	// Unregistered routes exercise middleware without touching the database.
	for range 1000 {
		request := httptest.NewRequest(http.MethodGet, "/api/not-found", nil)
		request.Header.Set("Origin", origin)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if response.Code != http.StatusTooManyRequests {
			continue
		}
		if got := response.Header().Get("Access-Control-Allow-Origin"); got != origin {
			t.Fatalf("429 CORS origin = %q, want %q", got, origin)
		}
		if got := response.Header().Get("Access-Control-Allow-Credentials"); got != "true" {
			t.Fatalf("429 allow credentials = %q, want true", got)
		}

		preflight := httptest.NewRequest(http.MethodOptions, "/api/not-found", nil)
		preflight.Header.Set("Origin", origin)
		preflight.Header.Set("Access-Control-Request-Method", http.MethodPost)
		response = httptest.NewRecorder()
		handler.ServeHTTP(response, preflight)
		if response.Code != http.StatusNoContent {
			t.Fatalf("preflight status = %d, want 204", response.Code)
		}
		return
	}
	t.Fatal("rate limit was never enforced")
}
