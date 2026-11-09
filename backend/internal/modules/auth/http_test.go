package auth_test

import (
	"backend/internal/modules/auth"
	"backend/internal/modules/session"
	"backend/internal/modules/user"
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

type authServiceStub struct {
	register func(context.Context, auth.RegisterDTO) (user.User, error)
	login    func(context.Context, auth.LoginDTO) (user.User, session.Session, string, error)
	me       func(context.Context, string) (user.User, error)
	logout   func(context.Context, string) error
}

func (s authServiceStub) Register(ctx context.Context, p auth.RegisterDTO) (user.User, error) {
	return s.register(ctx, p)
}
func (s authServiceStub) Login(ctx context.Context, p auth.LoginDTO) (user.User, session.Session, string, error) {
	return s.login(ctx, p)
}
func (s authServiceStub) Me(ctx context.Context, token string) (user.User, error) {
	return s.me(ctx, token)
}
func (s authServiceStub) Logout(ctx context.Context, token string) error { return s.logout(ctx, token) }

func authHandler(svc authServiceStub, secure bool) http.Handler {
	mux := http.NewServeMux()
	auth.NewHandler(svc, slog.New(slog.NewTextHandler(io.Discard, nil)), secure).Register(mux)
	return mux
}

func TestHandler_RejectsInvalidPayload(t *testing.T) {
	t.Parallel()
	for _, path := range []string{"/auth/register", "/auth/login"} {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			for _, tt := range []struct {
				name string
				body string
				code string
			}{
				{"empty", "", "INVALID_JSON"},
				{"malformed", "{", "INVALID_JSON"},
				{"unknown field", `{"unexpected":true}`, "INVALID_JSON"},
				{"multiple values", `{} {}`, "INVALID_JSON"},
				{"trailing garbage", `{} x`, "INVALID_JSON"},
				{"oversized", `{"email":"` + strings.Repeat("a", 1<<20) + `"}`, "INVALID_JSON"},
				{"missing required fields", `{}`, "VALIDATION_ERROR"},
				{"null", `null`, "VALIDATION_ERROR"},
			} {
				t.Run(tt.name, func(t *testing.T) {
					t.Parallel()
					w := httptest.NewRecorder()
					authHandler(authServiceStub{}, false).ServeHTTP(w, httptest.NewRequest(http.MethodPost, path, strings.NewReader(tt.body)))
					assertHTTPError(t, w, http.StatusBadRequest, tt.code, "")
				})
			}
		})
	}
}

func TestHandler_ErrorResponses(t *testing.T) {
	t.Parallel()
	failure := errors.New("secret database credentials")
	for _, tt := range []struct {
		name    string
		method  string
		path    string
		cookie  bool
		svcErr  error
		status  int
		code    string
		message string
	}{
		{"duplicate email", "POST", "/auth/register", false, user.ErrEmailTaken, 409, "EMAIL_TAKEN", user.ErrEmailTaken.Error()},
		{"registration failed", "POST", "/auth/register", false, failure, 500, "INTERNAL_ERROR", "internal error"},
		{"invalid credentials", "POST", "/auth/login", false, auth.ErrInvalidCredentials, 401, "INVALID_CREDENTIALS", auth.ErrInvalidCredentials.Error()},
		{"login failed", "POST", "/auth/login", false, failure, 500, "INTERNAL_ERROR", "internal error"},
		{"missing cookie", "GET", "/auth/me", false, nil, 401, "UNAUTHENTICATED", auth.ErrUnauthenticated.Error()},
		{"invalid session", "GET", "/auth/me", true, auth.ErrUnauthenticated, 401, "UNAUTHENTICATED", auth.ErrUnauthenticated.Error()},
		{"current user failed", "GET", "/auth/me", true, failure, 500, "INTERNAL_ERROR", "internal error"},
		{"logout failed", "POST", "/auth/logout", true, failure, 500, "INTERNAL_ERROR", "internal error"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := authServiceStub{
				register: func(context.Context, auth.RegisterDTO) (user.User, error) { return user.User{}, tt.svcErr },
				login: func(context.Context, auth.LoginDTO) (user.User, session.Session, string, error) {
					return user.User{}, session.Session{}, "", tt.svcErr
				},
				me: func(context.Context, string) (user.User, error) {
					if !tt.cookie {
						t.Fatal("request without a cookie reached the service")
					}
					return user.User{}, tt.svcErr
				},
				logout: func(context.Context, string) error { return tt.svcErr },
			}
			r := httptest.NewRequest(tt.method, tt.path, strings.NewReader(`{"name":"Alice","email":"alice@example.com","password":"password123"}`))
			if tt.path == "/auth/login" {
				r = httptest.NewRequest(tt.method, tt.path, strings.NewReader(`{"email":"alice@example.com","password":"password123"}`))
			}
			if tt.cookie {
				r.AddCookie(&http.Cookie{Name: "session", Value: "raw-token"})
			}
			w := httptest.NewRecorder()
			authHandler(svc, false).ServeHTTP(w, r)
			assertHTTPError(t, w, tt.status, tt.code, tt.message)
			if tt.path != "/auth/register" && w.Header().Get("Cache-Control") != "no-store" {
				t.Error("authentication response must not be cached")
			}
			if len(w.Result().Cookies()) != 0 {
				t.Error("failed request changed the session cookie")
			}
		})
	}
}

func TestHandler_SessionCookies(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name   string
		secure bool
	}{
		{"development", false},
		{"production", true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			expires := time.Date(2030, 1, 2, 3, 4, 5, 0, time.UTC)
			u := user.User{ID: "user-id", Name: "Alice", Email: "alice@example.com"}
			svc := authServiceStub{
				login: func(_ context.Context, p auth.LoginDTO) (user.User, session.Session, string, error) {
					if p.Email != u.Email || p.Password != " password " {
						t.Errorf("login payload = %+v, want trimmed email and original password", p)
					}
					return u, session.Session{ExpiresAt: expires}, "raw-token", nil
				},
				logout: func(_ context.Context, token string) error {
					if token != "raw-token" {
						t.Errorf("logout token = %q", token)
					}
					return nil
				},
			}
			h := authHandler(svc, tt.secure)
			w := httptest.NewRecorder()
			h.ServeHTTP(w, httptest.NewRequest("POST", "/auth/login", strings.NewReader(`{"email":" alice@example.com ","password":" password "}`)))
			if w.Code != http.StatusOK {
				t.Fatalf("login status = %d: %s", w.Code, w.Body.String())
			}
			var body struct{ Data user.User }
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil || body.Data != u {
				t.Fatalf("login response = %s (%v)", w.Body.String(), err)
			}
			for _, secret := range []string{"password", "token", "session", "raw-token"} {
				if strings.Contains(w.Body.String(), secret) {
					t.Errorf("login JSON exposes %q", secret)
				}
			}
			cookies := w.Result().Cookies()
			if len(cookies) != 1 {
				t.Fatalf("login cookies = %v, want one session cookie", cookies)
			}
			cookie := cookies[0]
			if cookie.Name != "session" || cookie.Value != "raw-token" || cookie.Path != "/" || !cookie.HttpOnly || cookie.Secure != tt.secure || cookie.SameSite != http.SameSiteLaxMode || cookie.MaxAge != 86400 || !cookie.Expires.Equal(expires) {
				t.Errorf("incorrect login cookie: %+v", cookie)
			}
			if w.Header().Get("Cache-Control") != "no-store" {
				t.Error("login response must not be cached")
			}

			for _, withCookie := range []bool{true, false} {
				r := httptest.NewRequest("POST", "/auth/logout", nil)
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

func assertHTTPError(t *testing.T, w *httptest.ResponseRecorder, status int, code, message string) {
	t.Helper()
	var body struct {
		Error struct{ Code, Message string }
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid error JSON: %v", err)
	}
	if w.Code != status || body.Error.Code != code || (message != "" && body.Error.Message != message) {
		t.Errorf("response = %d %s, want %d %s %q", w.Code, w.Body.String(), status, code, message)
	}
	if w.Header().Get("Content-Type") != "application/json" {
		t.Error("error response must have JSON content type")
	}
}
