package app_test

import (
	"io"
	"log/slog"
	"net/http/httptest"
	"testing"
	"testing/synctest"

	"backend/internal/app"
	"backend/internal/config"
)

func TestGlobalRateLimit(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		const origin = "http://localhost:3000"
		handler, err := app.New(config.Config{
			Env: "test", CORSOrigin: origin, RateLimitRPS: 1, RateLimitBurst: 2,
		}, nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
		if err != nil {
			t.Fatal(err)
		}
		for _, tt := range []struct {
			method, path string
			want         int
		}{
			{"GET", "/outside", 404},
			{"OPTIONS", "/api/auth/login", 204},
			{"GET", "/api/missing-a", 404},
			{"POST", "/api/missing-b", 404},
			{"GET", "/api/auth/me", 429},
			{"OPTIONS", "/api/auth/register", 204},
			{"POST", "/api/auth/login", 429},
			{"GET", "/outside", 404},
		} {
			req := httptest.NewRequest(tt.method, tt.path, nil)
			req.RemoteAddr = "192.0.2.1:1234"
			if tt.method == "POST" {
				req.RemoteAddr = "192.0.2.2:5678"
			}
			req.Header.Set("Origin", origin)
			if tt.method == "OPTIONS" {
				req.Header.Set("Access-Control-Request-Method", "POST")
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != tt.want {
				t.Fatalf("%s %s: status = %d, want %d", tt.method, tt.path, rec.Code, tt.want)
			}
			if rec.Header().Get("Access-Control-Allow-Origin") != origin || rec.Header().Get("Access-Control-Allow-Credentials") != "true" {
				t.Fatalf("CORS headers missing: %v", rec.Header())
			}
		}
	})
}
