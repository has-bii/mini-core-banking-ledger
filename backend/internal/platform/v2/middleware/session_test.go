package middleware_test

import (
	"backend/internal/platform/v2/httpx"
	"backend/internal/platform/v2/middleware"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserIDFromContext(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name          string
		authenticated bool
		wantID        string
		wantOK        bool
	}{
		{name: "missing user ID"},
		{name: "authenticated user", authenticated: true, wantID: "user-123", wantOK: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// A string key with a similar name must not impersonate middleware authentication.
			ctx := context.WithValue(context.Background(), "userID", "untrusted-user")
			request := httptest.NewRequest(http.MethodGet, "/private", nil).WithContext(ctx)
			if tt.authenticated {
				request.AddCookie(&http.Cookie{Name: "session", Value: "session-token"})
			}
			nextCalls := 0
			handler := middleware.GetUserFromSession(func(context.Context, string) (string, error) {
				return "user-123", nil
			})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalls++
				if id, ok := middleware.UserIDFromContext(r.Context()); id != tt.wantID || ok != tt.wantOK {
					t.Errorf("UserIDFromContext = (%q, %t), want (%q, %t)", id, ok, tt.wantID, tt.wantOK)
				}
			}))

			handler.ServeHTTP(httptest.NewRecorder(), request)

			if nextCalls != 1 {
				t.Errorf("next calls = %d, want 1", nextCalls)
			}
			if id, ok := middleware.UserIDFromContext(request.Context()); id != "" || ok {
				t.Errorf("original request was authenticated: (%q, %t)", id, ok)
			}
		})
	}
}

func TestGetUserFromSession(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name        string
		cookie      *http.Cookie
		userID      string
		wantLookups int
	}{
		{name: "no cookie"},
		{name: "unrelated cookie", cookie: &http.Cookie{Name: "other", Value: "token"}},
		{name: "empty session", cookie: &http.Cookie{Name: "session"}},
		{name: "unknown session", cookie: &http.Cookie{Name: "session", Value: "unknown-token"}, wantLookups: 1},
		{name: "valid session", cookie: &http.Cookie{Name: "session", Value: "session-token"}, userID: "user-123", wantLookups: 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			type contextKey struct{}
			ctx := context.WithValue(context.Background(), contextKey{}, "request-value")
			request := httptest.NewRequest(http.MethodGet, "/private", nil).WithContext(ctx)
			if tt.cookie != nil {
				request.AddCookie(tt.cookie)
			}
			recorder := httptest.NewRecorder()
			lookupCalls, nextCalls := 0, 0
			handler := middleware.GetUserFromSession(func(ctx context.Context, token string) (string, error) {
				lookupCalls++
				if ctx != request.Context() {
					t.Error("lookup did not receive original request context")
				}
				if tt.cookie == nil || token != tt.cookie.Value {
					t.Errorf("unexpected lookup token: %q", token)
				}
				return tt.userID, nil
			})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalls++
				if id, ok := middleware.UserIDFromContext(r.Context()); id != tt.userID || ok != (tt.userID != "") {
					t.Errorf("downstream user = (%q, %t), want (%q, %t)", id, ok, tt.userID, tt.userID != "")
				}
				if r.Context().Value(contextKey{}) != "request-value" {
					t.Error("downstream lost original context value")
				}
				w.WriteHeader(http.StatusAccepted)
				_, _ = w.Write([]byte("downstream response"))
			}))

			handler.ServeHTTP(recorder, request)

			if lookupCalls != tt.wantLookups || nextCalls != 1 {
				t.Errorf("lookup calls = %d, next calls = %d; want %d, 1", lookupCalls, nextCalls, tt.wantLookups)
			}
			if recorder.Code != http.StatusAccepted || recorder.Body.String() != "downstream response" {
				t.Errorf("downstream response changed: status %d, body %q", recorder.Code, recorder.Body.String())
			}
			if id, ok := middleware.UserIDFromContext(request.Context()); id != "" || ok {
				t.Errorf("original request was authenticated: (%q, %t)", id, ok)
			}
		})
	}
}

func TestGetUserFromSession_LookupError(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name   string
		userID string
	}{
		{name: "error without user ID"},
		{name: "error with user ID", userID: "user-123"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			request := httptest.NewRequest(http.MethodGet, "/private", nil)
			request.AddCookie(&http.Cookie{Name: "session", Value: "session-token"})
			recorder := httptest.NewRecorder()
			lookupCalls, nextCalls := 0, 0
			handler := middleware.GetUserFromSession(func(ctx context.Context, token string) (string, error) {
				lookupCalls++
				if ctx != request.Context() || token != "session-token" {
					t.Error("lookup did not receive the request context and session token")
				}
				return tt.userID, errors.New("private database failure")
			})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalls++
				_, _ = w.Write([]byte("protected content"))
			}))

			handler.ServeHTTP(recorder, request)

			if lookupCalls != 1 || nextCalls != 0 {
				t.Errorf("lookup calls = %d, next calls = %d; want 1, 0", lookupCalls, nextCalls)
			}
			if recorder.Code != http.StatusInternalServerError {
				t.Errorf("status = %d, want 500", recorder.Code)
			}
			if got := recorder.Header().Get("Content-Type"); got != "application/problem+json" {
				t.Errorf("Content-Type = %q, want application/problem+json", got)
			}
			var body httpx.ErrorResponse
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode error response: %v", err)
			}
			if body.Type != "about:blank" || body.Title != "Internal Server Error" || body.Status != 500 ||
				body.Detail != "Unexpected error has occurred" || body.Instance != "/private" || body.Error != nil {
				t.Errorf("unexpected error response: %+v", body)
			}
		})
	}
}

func TestAuthGuard(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name       string
		cookie     bool
		session    string
		userID     string
		wantStatus int
		wantNext   int
	}{
		{name: "anonymous request", wantStatus: http.StatusUnauthorized},
		{name: "empty session cookie", cookie: true, wantStatus: http.StatusUnauthorized},
		{name: "unknown session", cookie: true, session: "unknown-token", wantStatus: http.StatusUnauthorized},
		{name: "authenticated request", cookie: true, session: "session-token", userID: "user-123", wantStatus: http.StatusAccepted, wantNext: 1},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			request := httptest.NewRequest(http.MethodGet, "/private?secret=value", nil)
			if tt.cookie {
				request.AddCookie(&http.Cookie{Name: "session", Value: tt.session})
			}
			recorder := httptest.NewRecorder()
			nextCalls := 0
			handler := middleware.GetUserFromSession(func(context.Context, string) (string, error) {
				return tt.userID, nil
			})(middleware.AuthGuard(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalls++
				if id, ok := middleware.UserIDFromContext(r.Context()); !ok || id != tt.userID {
					t.Errorf("downstream user = (%q, %t), want (%q, true)", id, ok, tt.userID)
				}
				w.WriteHeader(http.StatusAccepted)
				_, _ = w.Write([]byte("protected content"))
			})))

			handler.ServeHTTP(recorder, request)

			if recorder.Code != tt.wantStatus || nextCalls != tt.wantNext {
				t.Errorf("status = %d, next calls = %d; want %d, %d", recorder.Code, nextCalls, tt.wantStatus, tt.wantNext)
			}
			if tt.wantNext == 1 {
				if recorder.Body.String() != "protected content" {
					t.Errorf("body = %q, want protected content", recorder.Body.String())
				}
				if got := recorder.Header().Get("Cache-Control"); got != "" {
					t.Errorf("authenticated Cache-Control = %q, want no middleware override", got)
				}
				return
			}
			if got := recorder.Header().Get("Cache-Control"); got != "no-store" {
				t.Errorf("Cache-Control = %q, want no-store", got)
			}
			if got := recorder.Header().Get("Content-Type"); got != "application/problem+json" {
				t.Errorf("Content-Type = %q, want application/problem+json", got)
			}
			var body httpx.ErrorResponse
			if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode error response: %v", err)
			}
			if body.Type != "about:blank" || body.Title != "Unauthorized" || body.Status != 401 ||
				body.Detail != "Session is expired or invalid" || body.Instance != "/private" || body.Error != nil {
				t.Errorf("unexpected unauthorized response: %+v", body)
			}
		})
	}
}
