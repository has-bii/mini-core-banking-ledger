package httpx

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestCORS(t *testing.T) {
	const allowed = "http://localhost:3000"
	cors, err := CORS(allowed)
	if err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct {
		name, method, origin, preflight string
		status                          int
		called, allow                   bool
	}{
		{"allowed read", "GET", allowed, "", 200, true, true},
		{"allowed write", "POST", allowed, "", 200, true, true},
		{"preflight", "OPTIONS", allowed, "POST", 204, false, true},
		{"untrusted read", "GET", "https://untrusted.example", "", 200, true, false},
		{"untrusted write", "POST", "https://untrusted.example", "", 403, false, false},
		{"untrusted preflight", "OPTIONS", "https://untrusted.example", "POST", 200, true, false},
		{"same origin", "POST", "http://example.com", "", 200, true, false},
		{"no origin", "POST", "", "", 200, true, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			handler := cors(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			}))
			req := httptest.NewRequest(tt.method, "http://example.com/api/auth/login", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			if tt.preflight != "" {
				req.Header.Set("Access-Control-Request-Method", tt.preflight)
				req.Header.Set("Access-Control-Request-Headers", "content-type")
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)
			if rec.Code != tt.status || called != tt.called {
				t.Fatalf("status=%d called=%v, want status=%d called=%v", rec.Code, called, tt.status, tt.called)
			}
			wantOrigin, wantCredentials := "", ""
			if tt.allow {
				wantOrigin, wantCredentials = allowed, "true"
			}
			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != wantOrigin {
				t.Errorf("allow origin=%q, want %q", got, wantOrigin)
			}
			if got := rec.Header().Get("Access-Control-Allow-Credentials"); got != wantCredentials {
				t.Errorf("allow credentials=%q, want %q", got, wantCredentials)
			}
			if rec.Header().Get("Vary") != "Origin" {
				t.Error("missing Vary: Origin")
			}
			if tt.preflight != "" && tt.allow {
				if !strings.Contains(rec.Header().Get("Access-Control-Allow-Methods"), "POST") || rec.Header().Get("Access-Control-Allow-Headers") != "Content-Type" {
					t.Error("preflight must allow POST and Content-Type")
				}
			}
		})
	}
}

func TestCORSRejectsInvalidOrigin(t *testing.T) {
	for _, origin := range []string{"", "*", "null", "localhost:3000", "http://localhost:3000/", "http://localhost:3000?query=yes"} {
		if _, err := CORS(origin); err == nil {
			t.Errorf("CORS(%q) should fail", origin)
		}
	}
}
