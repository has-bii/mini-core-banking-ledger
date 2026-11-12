package middleware_test

import (
	"backend/internal/platform/transport/httpx/middleware"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestGetUser(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name       string
		cookie     *http.Cookie
		userID     string
		lookupErr  error
		wantLookup bool
		wantUserID string
		wantStatus int
	}{
		{name: "missing cookie", wantStatus: http.StatusNoContent},
		{name: "empty cookie", cookie: &http.Cookie{Name: "session"}, wantStatus: http.StatusNoContent},
		{name: "valid identity", cookie: &http.Cookie{Name: "session", Value: "token"}, userID: "user-id", wantLookup: true, wantUserID: "user-id", wantStatus: http.StatusNoContent},
		{name: "anonymous identity", cookie: &http.Cookie{Name: "session", Value: "token"}, wantLookup: true, wantStatus: http.StatusNoContent},
		{name: "database error", cookie: &http.Cookie{Name: "session", Value: "token"}, lookupErr: errors.New("private database details"), wantLookup: true, wantStatus: http.StatusInternalServerError},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			request := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(t.Context())
			if tt.cookie != nil {
				request.AddCookie(tt.cookie)
			}

			lookupCalled := false
			lookupUserID := func(ctx context.Context, token string) (string, error) {
				lookupCalled = true
				if ctx != request.Context() || token != "token" {
					t.Fatal("lookup did not receive the request context and cookie token")
				}
				return tt.userID, tt.lookupErr
			}

			nextCalled := false
			handler := middleware.GetUser(lookupUserID)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				id, ok := middleware.UserIDFromContext(r.Context())
				if id != tt.wantUserID || ok != (tt.wantUserID != "") {
					t.Errorf("user identity = %q, %v; want %q, %v", id, ok, tt.wantUserID, tt.wantUserID != "")
				}
				w.WriteHeader(http.StatusNoContent)
			}))

			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", response.Code, tt.wantStatus)
			}
			if lookupCalled != tt.wantLookup {
				t.Errorf("lookup called = %v, want %v", lookupCalled, tt.wantLookup)
			}
			if nextCalled != (tt.wantStatus == http.StatusNoContent) {
				t.Errorf("next called = %v for status %d", nextCalled, tt.wantStatus)
			}
			if tt.wantStatus == http.StatusInternalServerError {
				body := response.Body.String()
				if !strings.Contains(body, `"code":"INTERNAL_ERROR"`) || strings.Contains(body, tt.lookupErr.Error()) {
					t.Errorf("unexpected error response: %s", body)
				}
			}
			if _, ok := middleware.UserIDFromContext(request.Context()); ok {
				t.Error("middleware modified the original request context")
			}
		})
	}
}

func TestAuthGuard(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name       string
		cookie     bool
		userID     string
		wantStatus int
	}{
		{name: "missing identity", wantStatus: http.StatusUnauthorized},
		{name: "unverified cookie with empty identity", cookie: true, wantStatus: http.StatusUnauthorized},
		{name: "verified identity", cookie: true, userID: "user-id", wantStatus: http.StatusNoContent},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			lookupUserID := func(context.Context, string) (string, error) {
				return tt.userID, nil
			}
			nextCalled := false
			handler := middleware.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				if id, ok := middleware.UserIDFromContext(r.Context()); !ok || id != tt.userID {
					t.Errorf("downstream identity = %q, %v; want %q, true", id, ok, tt.userID)
				}
				w.WriteHeader(http.StatusNoContent)
			}), middleware.GetUser(lookupUserID), middleware.AuthGuard)

			request := httptest.NewRequest(http.MethodGet, "/private", nil)
			if tt.cookie {
				request.AddCookie(&http.Cookie{Name: "session", Value: "token"})
			}
			response := httptest.NewRecorder()
			handler.ServeHTTP(response, request)

			if response.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", response.Code, tt.wantStatus)
			}
			if nextCalled != (tt.wantStatus == http.StatusNoContent) {
				t.Errorf("next called = %v for status %d", nextCalled, tt.wantStatus)
			}
			if tt.wantStatus == http.StatusUnauthorized {
				if response.Header().Get("Cache-Control") != "no-store" {
					t.Error("unauthorized response must not be cached")
				}
				var body struct {
					Error struct {
						Code    string `json:"code"`
						Message string `json:"message"`
					} `json:"error"`
				}
				if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
					t.Fatal(err)
				}
				if body.Error.Code != "UNAUTHENTICATED" || body.Error.Message != "authentication required" {
					t.Errorf("unexpected error response: %+v", body.Error)
				}
			}
		})
	}
}
