//go:build integration

package auth_test

import (
	"backend/internal/modules/account"
	"backend/internal/modules/auth"
	"backend/internal/modules/session"
	"backend/internal/modules/user"
	"backend/internal/platform/transport/httpx/middleware"
	"backend/internal/testutil"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func realAuthHandler(t *testing.T, secure bool) (http.Handler, *auth.Service, *session.Service) {
	t.Helper()
	pool := testutil.Postgres(t)
	sessions := session.NewService(session.NewStore(pool))
	svc := auth.NewService(pool, user.NewStore(pool), account.NewStore(pool), sessions)
	mux := http.NewServeMux()
	auth.NewHandler(svc, slog.New(slog.NewTextHandler(io.Discard, nil)), secure).Register(mux)
	h := middleware.GetUser(func(ctx context.Context, token string) (string, error) {
		current, err := sessions.FindByToken(ctx, token)
		if errors.Is(err, session.ErrNotFound) {
			return "", nil
		}
		return current.UserID, err
	})(mux)
	return h, svc, sessions
}

func TestHandler_DomainErrorResponses(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, path, payload, code, message string
		status                             int
	}{
		{"duplicate email", "/auth/register", `{"name":"Alice","email":"alice@example.com","password":"password123"}`, "EMAIL_TAKEN", "email already taken", http.StatusConflict},
		{"invalid credentials", "/auth/login", `{"email":"alice@example.com","password":"incorrect"}`, "INVALID_CREDENTIALS", "invalid email or password", http.StatusUnauthorized},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h, svc, _ := realAuthHandler(t, false)
			if _, err := svc.Register(t.Context(), auth.RegisterDTO{Name: "Alice", Email: "alice@example.com", Password: "password123"}); err != nil {
				t.Fatal(err)
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest("POST", tt.path, strings.NewReader(tt.payload)).WithContext(t.Context()))
			assertHTTPError(t, w, tt.status, tt.code, tt.message)
			if tt.path == "/auth/login" && w.Header().Get("Cache-Control") != "no-store" {
				t.Error("failed login must not be cached")
			}
			if len(w.Result().Cookies()) != 0 {
				t.Error("failed request changed session cookies")
			}
		})
	}
}

func TestHandler_SessionCookiesAndResponses(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name   string
		secure bool
	}{{"development", false}, {"production", true}} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			h, _, sessions := realAuthHandler(t, tt.secure)
			registered := httptest.NewRecorder()
			h.ServeHTTP(registered, httptest.NewRequest("POST", "/auth/register", strings.NewReader(`{"name":"Alice","email":"alice@example.com","password":" password "}`)).WithContext(t.Context()))
			if registered.Code != http.StatusCreated || len(registered.Result().Cookies()) != 0 {
				t.Fatalf("registration = %d %s; must not create a cookie", registered.Code, registered.Body.String())
			}
			var registration struct{ Data user.User }
			if err := json.Unmarshal(registered.Body.Bytes(), &registration); err != nil || registration.Data.ID == "" {
				t.Fatalf("registration response = %s (%v)", registered.Body.String(), err)
			}
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest("POST", "/auth/login", strings.NewReader(`{"email":" alice@example.com ","password":" password "}`)).WithContext(t.Context()))
			if w.Code != http.StatusOK {
				t.Fatalf("login status = %d: %s", w.Code, w.Body.String())
			}
			var body struct{ Data user.User }
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil || body.Data != registration.Data {
				t.Fatalf("login response = %s (%v)", w.Body.String(), err)
			}
			cookies := w.Result().Cookies()
			if len(cookies) != 1 {
				t.Fatalf("login cookies = %v, want one session cookie", cookies)
			}
			cookie := cookies[0]
			current, err := sessions.FindByToken(t.Context(), cookie.Value)
			if err != nil {
				t.Fatal(err)
			}
			if cookie.Name != "session" || cookie.Value == "" || cookie.Path != "/" || !cookie.HttpOnly || cookie.Secure != tt.secure || cookie.SameSite != http.SameSiteLaxMode || cookie.MaxAge != 86400 || !cookie.Expires.Equal(current.ExpiresAt.Truncate(time.Second)) {
				t.Errorf("incorrect login cookie: %+v", cookie)
			}
			if w.Header().Get("Cache-Control") != "no-store" {
				t.Error("login response must not be cached")
			}
			me := httptest.NewRecorder()
			r := httptest.NewRequest("GET", "/auth/me", nil).WithContext(t.Context())
			r.AddCookie(cookie)
			h.ServeHTTP(me, r)
			if me.Code != http.StatusOK || me.Header().Get("Cache-Control") != "no-store" {
				t.Fatalf("current user = %d %s", me.Code, me.Body.String())
			}
			if err := json.Unmarshal(me.Body.Bytes(), &body); err != nil || body.Data != registration.Data {
				t.Fatalf("current user response = %s (%v)", me.Body.String(), err)
			}
			for _, response := range []*httptest.ResponseRecorder{registered, w, me} {
				for _, secret := range []string{"password", "token", "session", cookie.Value} {
					if strings.Contains(response.Body.String(), secret) {
						t.Errorf("user JSON exposes %q", secret)
					}
				}
			}
			for _, withCookie := range []bool{true, false} {
				r := httptest.NewRequest("POST", "/auth/logout", nil).WithContext(t.Context())
				if withCookie {
					r.AddCookie(cookie)
				}
				w := httptest.NewRecorder()
				h.ServeHTTP(w, r)
				if w.Code != http.StatusNoContent || w.Body.Len() != 0 || w.Header().Get("Cache-Control") != "no-store" {
					t.Fatalf("logout with cookie=%v: status=%d, body=%s", withCookie, w.Code, w.Body.String())
				}
				cookies := w.Result().Cookies()
				if len(cookies) != 1 {
					t.Fatalf("logout cookies = %v, want one expired cookie", cookies)
				}
				cleared := cookies[0]
				if cleared.Name != "session" || cleared.Value != "" || cleared.MaxAge != -1 || !cleared.Expires.Before(time.Now()) || cleared.Path != "/" || !cleared.HttpOnly || cleared.Secure != tt.secure || cleared.SameSite != http.SameSiteLaxMode {
					t.Errorf("logout did not securely clear the cookie: %+v", cleared)
				}
			}
		})
	}
}
