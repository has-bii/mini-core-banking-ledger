package auth_test

import (
	"backend/internal/modules/auth"
	"backend/internal/modules/session"
	"backend/internal/modules/user"
	"backend/internal/platform/transport/httpx/middleware"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Failure injection only: ordinary HTTP outcomes use real services and PostgreSQL.
type authServiceStub struct{ err error }

func (s authServiceStub) Register(context.Context, auth.RegisterDTO) (user.User, error) {
	return user.User{}, s.err
}
func (s authServiceStub) Login(context.Context, auth.LoginDTO) (user.User, session.Session, string, error) {
	return user.User{}, session.Session{}, "", s.err
}
func (s authServiceStub) Me(context.Context, string) (user.User, error) { return user.User{}, s.err }
func (s authServiceStub) Logout(context.Context, string) error          { return s.err }

func TestHandler_RejectsInvalidPayload(t *testing.T) {
	t.Parallel()
	for _, path := range []string{"/auth/register", "/auth/login"} {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			for _, tt := range []struct{ name, body, code string }{
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
					mux := http.NewServeMux()
					// Nil dependencies prove parsing/validation stops before business logic.
					auth.NewHandler(nil, slog.New(slog.NewTextHandler(io.Discard, nil)), false).Register(mux)
					w := httptest.NewRecorder()
					mux.ServeHTTP(w, httptest.NewRequest(http.MethodPost, path, strings.NewReader(tt.body)))
					assertHTTPError(t, w, http.StatusBadRequest, tt.code, "")
				})
			}
		})
	}
}

func TestHandler_UnexpectedErrors(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, method, path, payload string }{
		{"registration", "POST", "/auth/register", `{"name":"Alice","email":"alice@example.com","password":"password123"}`},
		{"login", "POST", "/auth/login", `{"email":"alice@example.com","password":"password123"}`},
		{"current user", "GET", "/auth/me", ""},
		{"logout", "POST", "/auth/logout", ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mux := http.NewServeMux()
			auth.NewHandler(authServiceStub{err: errors.New("secret database credentials")}, slog.New(slog.NewTextHandler(io.Discard, nil)), false).Register(mux)
			// Resolve identity only to reach the handler's injected failure.
			h := middleware.GetUser(func(context.Context, string) (string, error) { return "user-id", nil })(mux)
			r := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.payload))
			r.AddCookie(&http.Cookie{Name: "session", Value: "raw-token"})
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			assertHTTPError(t, w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal error")
			if tt.path != "/auth/register" && w.Header().Get("Cache-Control") != "no-store" {
				t.Error("authentication response must not be cached")
			}
			if len(w.Result().Cookies()) != 0 {
				t.Error("failed request changed the session cookie")
			}
		})
	}
}

func TestHandler_MeRequiresResolvedIdentity(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	auth.NewHandler(nil, slog.New(slog.NewTextHandler(io.Discard, nil)), false).Register(mux)
	for _, name := range []string{"missing cookie", "unverified cookie"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			r := httptest.NewRequest(http.MethodGet, "/auth/me", nil)
			if name == "unverified cookie" {
				r.AddCookie(&http.Cookie{Name: "session", Value: "raw-token"})
			}
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)
			assertHTTPError(t, w, http.StatusUnauthorized, "UNAUTHENTICATED", "authentication required")
			if w.Header().Get("Cache-Control") != "no-store" || len(w.Result().Cookies()) != 0 {
				t.Error("unauthorized response must not be cached or change cookies")
			}
		})
	}
}

// A user disappearing after session resolution is a race, not a normal fixture state.
func TestHandler_DeletedUserAfterResolution(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	auth.NewHandler(authServiceStub{err: auth.ErrUnauthenticated}, slog.New(slog.NewTextHandler(io.Discard, nil)), false).Register(mux)
	h := middleware.GetUser(func(context.Context, string) (string, error) { return "deleted-user", nil })(mux)
	r := httptest.NewRequest("GET", "/auth/me", nil)
	r.AddCookie(&http.Cookie{Name: "session", Value: "raw-token"})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	assertHTTPError(t, w, http.StatusUnauthorized, "UNAUTHENTICATED", "authentication required")
	if w.Header().Get("Cache-Control") != "no-store" || len(w.Result().Cookies()) != 0 {
		t.Error("unauthorized response must not be cached or change cookies")
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
