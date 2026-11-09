package app_test

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"backend/internal/app"
)

func TestAPIPrefix(t *testing.T) {
	handler := app.New("test", nil, slog.New(slog.NewTextHandler(io.Discard, nil)))
	for _, tc := range []struct {
		path   string
		status int
	}{
		{"/api/auth/me", http.StatusUnauthorized},
		{"/auth/me", http.StatusNotFound},
	} {
		t.Run(tc.path, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tc.path, nil))
			if rec.Code != tc.status {
				t.Fatalf("status = %d, want %d", rec.Code, tc.status)
			}
		})
	}
}
