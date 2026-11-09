package auth_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"backend/internal/modules/auth"
	"backend/internal/modules/session"
	"backend/internal/modules/user"
)

type stubAuth struct {
	register func(context.Context, auth.RegisterDTO) (user.User, error)
	login    func(context.Context, auth.LoginDTO) (user.User, session.Session, string, error)
	me       func(context.Context, string) (user.User, error)
	logout   func(context.Context, string) error
}

func (s stubAuth) Register(ctx context.Context, p auth.RegisterDTO) (user.User, error) {
	return s.register(ctx, p)
}
func (s stubAuth) Login(ctx context.Context, p auth.LoginDTO) (user.User, session.Session, string, error) {
	return s.login(ctx, p)
}
func (s stubAuth) Me(ctx context.Context, token string) (user.User, error) {
	return s.me(ctx, token)
}
func (s stubAuth) Logout(ctx context.Context, token string) error {
	return s.logout(ctx, token)
}

func authHandler(svc stubAuth, secure bool) http.Handler {
	mux := http.NewServeMux()
	auth.NewHandler(svc, slog.New(slog.NewTextHandler(io.Discard, nil)), secure).Register(mux)
	return mux
}

func serveAuth(t *testing.T, h http.Handler, method, path, body string, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body)).WithContext(t.Context())
	req.Header.Set("Content-Type", "application/json")
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func assertHTTPError(t *testing.T, rec *httptest.ResponseRecorder, status int, code string) {
	t.Helper()
	if rec.Code != status {
		t.Fatalf("status = %d, want %d; body = %s", rec.Code, status, rec.Body.String())
	}
	var body struct {
		Error struct{ Code, Message string }
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error.Code != code || body.Error.Message == "" {
		t.Errorf("error = %+v, want code %s and a message", body.Error, code)
	}
	if rec.Header().Get("Content-Type") != "application/json" {
		t.Error("error response must be JSON")
	}
}

func TestAuthHTTPRejectsInvalidInput(t *testing.T) {
	for _, path := range []string{"/auth/register", "/auth/login"} {
		t.Run(path, func(t *testing.T) {
			for _, tc := range []struct{ name, body, code string }{
				{"empty", "", "INVALID_JSON"},
				{"malformed", "{", "INVALID_JSON"},
				{"unknown field", `{"unexpected":true}`, "INVALID_JSON"},
				{"multiple values", `{} {}`, "INVALID_JSON"},
				{"wrong field type", `{"email":42}`, "INVALID_JSON"},
				{"oversized", `{"email":"` + strings.Repeat("a", 1<<20) + `"}`, "INVALID_JSON"},
				{"missing fields", `{}`, "VALIDATION_ERROR"},
				{"invalid email", `{"email":"invalid","password":"password123"}`, "VALIDATION_ERROR"},
				{"short password", `{"email":"alice@example.com","password":"short"}`, "VALIDATION_ERROR"},
			} {
				t.Run(tc.name, func(t *testing.T) {
					// No callbacks: invalid requests must not reach the auth module.
					rec := serveAuth(t, authHandler(stubAuth{}, false), http.MethodPost, path, tc.body, nil)
					assertHTTPError(t, rec, http.StatusBadRequest, tc.code)
					if len(rec.Result().Cookies()) != 0 {
						t.Error("invalid request set a cookie")
					}
				})
			}
		})
	}
}

func TestAuthHTTPMapsErrors(t *testing.T) {
	failure := errors.New("secret database connection details")
	for _, tc := range []struct {
		name, method, path, body, code string
		err                            error
		status                         int
	}{
		{"duplicate", "POST", "/auth/register", `{"name":"Alice","email":"alice@example.com","password":"password123"}`, "EMAIL_TAKEN", user.ErrEmailTaken, 409},
		{"register failure", "POST", "/auth/register", `{"name":"Alice","email":"alice@example.com","password":"password123"}`, "INTERNAL_ERROR", failure, 500},
		{"invalid credentials", "POST", "/auth/login", `{"email":"alice@example.com","password":"password123"}`, "INVALID_CREDENTIALS", auth.ErrInvalidCredentials, 401},
		{"login failure", "POST", "/auth/login", `{"email":"alice@example.com","password":"password123"}`, "INTERNAL_ERROR", failure, 500},
		{"unauthenticated", "GET", "/auth/me", "", "UNAUTHENTICATED", auth.ErrUnauthenticated, 401},
		{"me failure", "GET", "/auth/me", "", "INTERNAL_ERROR", failure, 500},
		{"logout failure", "POST", "/auth/logout", "", "INTERNAL_ERROR", failure, 500},
	} {
		t.Run(tc.name, func(t *testing.T) {
			wrapped := fmt.Errorf("operation: %w", tc.err)
			svc := stubAuth{
				register: func(context.Context, auth.RegisterDTO) (user.User, error) { return user.User{}, wrapped },
				login: func(context.Context, auth.LoginDTO) (user.User, session.Session, string, error) {
					return user.User{}, session.Session{}, "", wrapped
				},
				me:     func(context.Context, string) (user.User, error) { return user.User{}, wrapped },
				logout: func(context.Context, string) error { return wrapped },
			}
			rec := serveAuth(t, authHandler(svc, true), tc.method, tc.path, tc.body, &http.Cookie{Name: "session", Value: "raw-token"})
			assertHTTPError(t, rec, tc.status, tc.code)
			if tc.status == 500 && (strings.Contains(rec.Body.String(), failure.Error()) || strings.Contains(rec.Body.String(), "operation")) {
				t.Error("response leaked internal error details")
			}
			if len(rec.Result().Cookies()) != 0 {
				t.Error("failed operation must not set or clear a session cookie")
			}
			if tc.path != "/auth/register" && rec.Header().Get("Cache-Control") != "no-store" {
				t.Error("auth response must disable caching")
			}
		})
	}
}

func TestAuthHTTPRegister(t *testing.T) {
	want := user.User{ID: "user-1", Name: "Alice", Email: "alice@example.com"}
	svc := stubAuth{register: func(_ context.Context, p auth.RegisterDTO) (user.User, error) {
		if p.Name != want.Name || p.Email != want.Email || p.Password != "password123" {
			t.Fatalf("unexpected validated payload: %+v", p)
		}
		return want, nil
	}}
	rec := serveAuth(t, authHandler(svc, false), "POST", "/auth/register", `{"name":" Alice ","email":" alice@example.com ","password":"password123"}`, nil)
	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d; body = %s", rec.Code, rec.Body.String())
	}
	assertUserJSON(t, rec, want)
}

func assertUserJSON(t *testing.T, rec *httptest.ResponseRecorder, want user.User) {
	t.Helper()
	var body struct{ Data user.User }
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Data != want {
		t.Errorf("user = %+v, want %+v", body.Data, want)
	}
	for _, secret := range []string{"password", "token", "raw-token"} {
		if strings.Contains(rec.Body.String(), secret) {
			t.Errorf("user response contains %q", secret)
		}
	}
}

func TestAuthHTTPCookies(t *testing.T) {
	for _, secure := range []bool{false, true} {
		t.Run(fmt.Sprintf("secure=%t", secure), func(t *testing.T) {
			want := user.User{ID: "user-1", Email: "alice@example.com"}
			expires := time.Now().UTC().Truncate(time.Second).Add(session.Lifetime)
			svc := stubAuth{
				login: func(_ context.Context, p auth.LoginDTO) (user.User, session.Session, string, error) {
					if p.Email != want.Email || p.Password != "password123" {
						t.Fatalf("unexpected validated login: %+v", p)
					}
					return want, session.Session{ExpiresAt: expires}, "raw-token", nil
				},
				logout: func(_ context.Context, token string) error {
					if token != "raw-token" {
						t.Errorf("logout token = %q", token)
					}
					return nil
				},
			}
			h := authHandler(svc, secure)
			login := serveAuth(t, h, "POST", "/auth/login", `{"email":" alice@example.com ","password":"password123"}`, nil)
			if login.Code != 200 || login.Header().Get("Cache-Control") != "no-store" {
				t.Fatalf("unexpected login: %d, %v", login.Code, login.Header())
			}
			assertUserJSON(t, login, want)
			cookies := login.Result().Cookies()
			if len(cookies) != 1 {
				t.Fatalf("login cookies = %v", cookies)
			}
			cookie := cookies[0]
			if cookie.Name != "session" || cookie.Value != "raw-token" || cookie.Domain != "" || cookie.Path != "/" || !cookie.HttpOnly || cookie.Secure != secure || cookie.SameSite != http.SameSiteLaxMode || cookie.MaxAge != 86400 || !cookie.Expires.Equal(expires) {
				t.Fatalf("unexpected session cookie: %+v", cookie)
			}
			logout := serveAuth(t, h, "POST", "/auth/logout", "", cookie)
			assertClearedCookie(t, logout, secure)
		})
	}
}

func assertClearedCookie(t *testing.T, rec *httptest.ResponseRecorder, secure bool) {
	t.Helper()
	if rec.Code != http.StatusNoContent || rec.Body.Len() != 0 || rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("unexpected logout: status=%d, headers=%v, body=%s", rec.Code, rec.Header(), rec.Body.String())
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 {
		t.Fatalf("logout cookies = %v", cookies)
	}
	c := cookies[0]
	if c.Name != "session" || c.Value != "" || c.MaxAge != -1 || c.Expires.IsZero() || !c.Expires.Before(time.Now()) || c.Path != "/" || c.Domain != "" || !c.HttpOnly || c.Secure != secure || c.SameSite != http.SameSiteLaxMode {
		t.Errorf("unexpected cleared cookie: %+v", c)
	}
}

func TestAuthHTTPMissingSession(t *testing.T) {
	for _, cookie := range []*http.Cookie{nil, {Name: "session", Value: ""}} {
		rec := serveAuth(t, authHandler(stubAuth{}, false), "GET", "/auth/me", "", cookie)
		assertHTTPError(t, rec, http.StatusUnauthorized, "UNAUTHENTICATED")
		if rec.Header().Get("Cache-Control") != "no-store" {
			t.Error("unauthenticated response must disable caching")
		}
	}
	assertClearedCookie(t, serveAuth(t, authHandler(stubAuth{}, false), "POST", "/auth/logout", "", nil), false)
}

func TestAuthHTTPMe(t *testing.T) {
	want := user.User{ID: "user-1", Email: "alice@example.com"}
	svc := stubAuth{me: func(_ context.Context, token string) (user.User, error) {
		if token != "raw-token" {
			t.Errorf("session token = %q", token)
		}
		return want, nil
	}}
	rec := serveAuth(t, authHandler(svc, false), "GET", "/auth/me", "", &http.Cookie{Name: "session", Value: "raw-token"})
	if rec.Code != 200 || rec.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("unexpected me response: %d, %v", rec.Code, rec.Header())
	}
	assertUserJSON(t, rec, want)
}
