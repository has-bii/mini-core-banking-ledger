package app_test

import (
	"backend/internal/app"
	"backend/internal/config"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew_RateLimitCORS(t *testing.T) {
	t.Parallel()

	const origin = "http://localhost:3000"
	handler, err := app.New(config.Config{
		Env: "test", CORSOrigin: origin, RatelimitCapacity: 1, RatelimitRefillRate: 0.001,
	}, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}

	// Unregistered routes exercise middleware without touching the database.
	for i := range 2 {
		request := httptest.NewRequest(http.MethodGet, "/api/not-found", nil)
		request.Header.Set("Origin", origin)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, request)
		if i == 0 {
			if response.Code != http.StatusNotFound {
				t.Fatalf("first request status = %d, want 404", response.Code)
			}
			continue
		}
		if response.Code != http.StatusTooManyRequests {
			t.Fatalf("second request status = %d, want 429", response.Code)
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

func TestNew_InvalidSession(t *testing.T) {
	t.Parallel()

	handler, err := app.New(config.Config{
		Env: "test", CORSOrigin: "http://localhost:3000", RatelimitCapacity: 50, RatelimitRefillRate: 10,
	}, nil,
		slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	request.AddCookie(&http.Cookie{Name: "session", Value: "malformed"})
	response := httptest.NewRecorder()
	handler.ServeHTTP(response, request)

	var body struct {
		Error struct{ Code, Message string }
	}
	if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if response.Code != http.StatusUnauthorized || body.Error.Code != "UNAUTHENTICATED" || body.Error.Message != "authentication required" {
		t.Fatalf("invalid session: status=%d body=%s, want authentication required", response.Code, response.Body.String())
	}
}
