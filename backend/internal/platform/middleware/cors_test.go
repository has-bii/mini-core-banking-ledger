package middleware_test

import (
	"backend/internal/platform/middleware"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"slices"
	"strings"
	"testing"
	"time"
)

func TestCORS_Origins(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name    string
		origins []string
		wantErr bool
	}{
		{name: "no trusted origins"},
		{name: "multiple trusted origins", origins: []string{"https://app.example.com", "http://localhost:3000"}},
		{name: "wildcard", origins: []string{"*"}, wantErr: true},
		{name: "empty origin", origins: []string{""}, wantErr: true},
		{name: "missing scheme", origins: []string{"app.example.com"}, wantErr: true},
		{name: "missing host", origins: []string{"https://"}, wantErr: true},
		{name: "path", origins: []string{"https://app.example.com/path"}, wantErr: true},
		{name: "query", origins: []string{"https://app.example.com?key=value"}, wantErr: true},
		{name: "fragment", origins: []string{"https://app.example.com#section"}, wantErr: true},
		{name: "malformed URL", origins: []string{"https://%"}, wantErr: true},
		{name: "invalid origin after valid origin", origins: []string{"https://app.example.com", "*"}, wantErr: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mw, err := middleware.CORS(&middleware.CORSOptions{Origin: tt.origins})
			if tt.wantErr {
				if err == nil {
					t.Fatal("CORS returned no error for invalid origin")
				}
				if mw != nil {
					t.Error("CORS returned middleware for invalid origin")
				}
				if !strings.Contains(err.Error(), "invalid CORS origin") {
					t.Errorf("error = %q, want invalid CORS origin context", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("CORS: %v", err)
			}
			if mw == nil {
				t.Error("CORS returned nil middleware")
			}
		})
	}
}

func TestCORS_Requests(t *testing.T) {
	t.Parallel()

	const trusted = "https://app.example.com"
	for _, tt := range []struct {
		name        string
		method      string
		origin      string
		fetchSite   string
		credentials bool
		wantStatus  int
		wantOrigin  string
	}{
		{name: "trusted origin", method: http.MethodPost, origin: trusted, wantStatus: http.StatusCreated, wantOrigin: trusted},
		{name: "trusted origin with credentials", method: http.MethodPost, origin: trusted, credentials: true, wantStatus: http.StatusCreated, wantOrigin: trusted},
		{name: "second trusted origin with port", method: http.MethodPost, origin: "http://localhost:3000", wantStatus: http.StatusCreated, wantOrigin: "http://localhost:3000"},
		{name: "untrusted unsafe request", method: http.MethodPost, origin: "https://evil.example", credentials: true, wantStatus: http.StatusForbidden},
		{name: "origin must match exactly", method: http.MethodPost, origin: trusted + ".evil.example", wantStatus: http.StatusForbidden},
		{name: "different scheme", method: http.MethodPost, origin: "http://app.example.com", wantStatus: http.StatusForbidden},
		{name: "different port", method: http.MethodPost, origin: trusted + ":8443", wantStatus: http.StatusForbidden},
		{name: "untrusted GET", method: http.MethodGet, origin: "https://evil.example", wantStatus: http.StatusCreated},
		{name: "untrusted HEAD", method: http.MethodHead, origin: "https://evil.example", wantStatus: http.StatusCreated},
		{name: "same origin", method: http.MethodPost, origin: "https://api.example.com", wantStatus: http.StatusCreated},
		{name: "no origin", method: http.MethodPost, wantStatus: http.StatusCreated},
		{name: "cross site without origin", method: http.MethodPost, fetchSite: "cross-site", wantStatus: http.StatusForbidden},
		{name: "cross site trusted origin", method: http.MethodPost, origin: trusted, fetchSite: "cross-site", wantStatus: http.StatusCreated, wantOrigin: trusted},
		{name: "cross site untrusted origin", method: http.MethodPost, origin: "https://evil.example", fetchSite: "cross-site", wantStatus: http.StatusForbidden},
		{name: "same site is not same origin", method: http.MethodPost, origin: "https://other.example.com", fetchSite: "same-site", wantStatus: http.StatusForbidden},
		{name: "same origin fetch metadata", method: http.MethodPost, fetchSite: "same-origin", wantStatus: http.StatusCreated},
		{name: "user initiated fetch metadata", method: http.MethodPost, fetchSite: "none", wantStatus: http.StatusCreated},
		{name: "trusted preflight", method: http.MethodOptions, origin: trusted, credentials: true, wantStatus: http.StatusNoContent, wantOrigin: trusted},
		{name: "untrusted preflight", method: http.MethodOptions, origin: "https://evil.example", fetchSite: "cross-site", credentials: true, wantStatus: http.StatusNoContent},
		{name: "OPTIONS without origin", method: http.MethodOptions, wantStatus: http.StatusNoContent},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mw, err := middleware.CORS(&middleware.CORSOptions{
				Origin:         []string{trusted, "http://localhost:3000"},
				AllowedMethods: "GET, POST, OPTIONS",
				AllowedHeaders: "Content-Type, Authorization",
				Credentials:    tt.credentials,
			})
			if err != nil {
				t.Fatalf("CORS: %v", err)
			}
			calls := 0
			handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				calls++
				w.Header().Set("X-Handler", "preserved")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte("created"))
			}))
			req := httptest.NewRequest(tt.method, "https://api.example.com/resource", nil)
			if tt.origin != "" {
				req.Header.Set("Origin", tt.origin)
			}
			if tt.fetchSite != "" {
				req.Header.Set("Sec-Fetch-Site", tt.fetchSite)
			}
			if tt.method == http.MethodOptions {
				req.Header.Set("Access-Control-Request-Method", http.MethodPost)
				req.Header.Set("Access-Control-Request-Headers", "Authorization")
			}
			rr := httptest.NewRecorder()
			rr.Header().Add("Vary", "Accept-Encoding")
			handler.ServeHTTP(rr, req)
			if rr.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", rr.Code, tt.wantStatus)
			}
			if tt.wantStatus == http.StatusForbidden {
				if got := rr.Header().Get("Content-Type"); got != "application/problem+json" {
					t.Errorf("Content-Type = %q, want application/problem+json", got)
				}
				var problem map[string]any
				if err := json.Unmarshal(rr.Body.Bytes(), &problem); err != nil {
					t.Fatalf("decode problem details: %v; body = %q", err, rr.Body.String())
				}
				want := map[string]any{
					"type":     "about:blank",
					"title":    "Forbidden",
					"status":   float64(403),
					"detail":   "Cross-origin request denied.",
					"instance": "/resource",
				}
				if !reflect.DeepEqual(problem, want) {
					t.Errorf("problem details = %v, want %v", problem, want)
				}
			}
			wantCredentials := ""
			if tt.credentials && tt.wantOrigin != "" {
				wantCredentials = "true"
			}
			for header, want := range map[string]string{
				"Access-Control-Allow-Origin":      tt.wantOrigin,
				"Access-Control-Allow-Credentials": wantCredentials,
				"Access-Control-Allow-Methods":     "GET, POST, OPTIONS",
				"Access-Control-Allow-Headers":     "Content-Type, Authorization",
				"Access-Control-Max-Age":           "",
			} {
				if got := rr.Header().Get(header); got != want {
					t.Errorf("%s = %q, want %q", header, got, want)
				}
			}
			vary := rr.Header().Values("Vary")
			if !slices.Contains(vary, "Origin") || !slices.Contains(vary, "Accept-Encoding") {
				t.Errorf("Vary = %v, want Origin and preserved Accept-Encoding", vary)
			}
			wantCalls := 0
			if tt.wantStatus == http.StatusCreated {
				wantCalls = 1
				if rr.Body.String() != "created" || rr.Header().Get("X-Handler") != "preserved" {
					t.Errorf("downstream response not preserved: headers %v, body %q", rr.Header(), rr.Body.String())
				}
			}
			if calls != wantCalls {
				t.Errorf("handler calls = %d, want %d", calls, wantCalls)
			}
			if tt.method == http.MethodOptions && rr.Body.Len() != 0 {
				t.Errorf("preflight body = %q, want empty", rr.Body.String())
			}
		})
	}
}

func TestCORS_MaxAge(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name   string
		maxAge time.Duration
		want   string
	}{
		{name: "negative", maxAge: -time.Second},
		{name: "zero"},
		{name: "positive subsecond", maxAge: 500 * time.Millisecond, want: "0"},
		{name: "one second", maxAge: time.Second, want: "1"},
		{name: "fractional seconds truncated", maxAge: 1500 * time.Millisecond, want: "1"},
		{name: "ten minutes", maxAge: 10 * time.Minute, want: "600"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mw, err := middleware.CORS(&middleware.CORSOptions{MaxAge: tt.maxAge})
			if err != nil {
				t.Fatalf("CORS: %v", err)
			}
			handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			for _, method := range []string{http.MethodOptions, http.MethodGet} {
				rr := httptest.NewRecorder()
				handler.ServeHTTP(rr, httptest.NewRequest(method, "/", nil))
				want := ""
				if method == http.MethodOptions {
					want = tt.want
				}
				if got := rr.Header().Get("Access-Control-Max-Age"); got != want {
					t.Errorf("%s max-age = %q, want %q", method, got, want)
				}
			}
		})
	}
}
