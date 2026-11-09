package httpx_test

import (
	"backend/internal/modules/session"
	"backend/internal/platform/httpx"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

type sessionServiceStub struct {
	find func(context.Context, string) (session.Session, error)
}

func (s sessionServiceStub) FindByToken(ctx context.Context, token string) (session.Session, error) {
	return s.find(ctx, token)
}

func TestGetUser(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name       string
		cookie     *http.Cookie
		lookupErr  error
		wantLookup bool
		wantUserID string
		wantStatus int
	}{
		{name: "missing cookie", wantStatus: http.StatusNoContent},
		{name: "empty cookie", cookie: &http.Cookie{Name: "session"}, wantStatus: http.StatusNoContent},
		{name: "valid session", cookie: &http.Cookie{Name: "session", Value: "token"}, wantLookup: true, wantUserID: "user-id", wantStatus: http.StatusNoContent},
		{name: "invalid or expired session", cookie: &http.Cookie{Name: "session", Value: "token"}, lookupErr: session.ErrNotFound, wantLookup: true, wantStatus: http.StatusNoContent},
		{name: "wrapped session error", cookie: &http.Cookie{Name: "session", Value: "token"}, lookupErr: fmt.Errorf("lookup: %w", session.ErrNotFound), wantLookup: true, wantStatus: http.StatusNoContent},
		{name: "database error", cookie: &http.Cookie{Name: "session", Value: "token"}, lookupErr: errors.New("private database details"), wantLookup: true, wantStatus: http.StatusInternalServerError},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			request := httptest.NewRequest(http.MethodGet, "/", nil).WithContext(t.Context())
			if tt.cookie != nil {
				request.AddCookie(tt.cookie)
			}

			lookupCalled := false
			svc := sessionServiceStub{find: func(ctx context.Context, token string) (session.Session, error) {
				lookupCalled = true
				if ctx != request.Context() || token != "token" {
					t.Fatal("lookup did not receive the request context and cookie token")
				}
				return session.Session{UserID: "user-id"}, tt.lookupErr
			}}

			nextCalled := false
			handler := httpx.GetUser(svc)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				id, ok := httpx.UserIDFromContext(r.Context())
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
			if _, ok := httpx.UserIDFromContext(request.Context()); ok {
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
		lookupErr  error
		wantStatus int
	}{
		{name: "missing identity", wantStatus: http.StatusUnauthorized},
		{name: "unverified cookie", cookie: true, lookupErr: session.ErrNotFound, wantStatus: http.StatusUnauthorized},
		{name: "empty identity", cookie: true, wantStatus: http.StatusUnauthorized},
		{name: "verified identity", cookie: true, userID: "user-id", wantStatus: http.StatusNoContent},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := sessionServiceStub{find: func(context.Context, string) (session.Session, error) {
				return session.Session{UserID: tt.userID}, tt.lookupErr
			}}
			nextCalled := false
			handler := httpx.Chain(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				if id, ok := httpx.UserIDFromContext(r.Context()); !ok || id != tt.userID {
					t.Errorf("downstream identity = %q, %v; want %q, true", id, ok, tt.userID)
				}
				w.WriteHeader(http.StatusNoContent)
			}), httpx.GetUser(svc), httpx.AuthGuard())

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
