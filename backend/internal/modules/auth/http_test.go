package auth_test

import (
	"backend/internal/modules/auth"
	"backend/internal/modules/session"
	"backend/internal/modules/user"
	"backend/internal/platform/middleware"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
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
			for _, tt := range []struct {
				name, body, title, message string
				status                     int
			}{
				{"malformed", "{", "Bad Request", "body contains badly-formed JSON", 400},
				{"empty", "", "Bad Request", "body must not be empty", 400},
				{"non-object", `null`, "Bad Request", "body must be a JSON object", 400},
				{"unknown field", `{"extra":true}`, "Bad Request", `body contains unknown field "extra"`, 400},
				{"wrong field type", `{"email":123}`, "Bad Request", `body contains an invalid value for field "email" (expected string)`, 400},
				{"multiple objects", `{} {}`, "Bad Request", "body must contain only one JSON object", 400},
				{"oversized trailing whitespace", `{}` + strings.Repeat(" ", 1<<20), "Request Entity Too Large", "body must not be larger than 1048576 bytes", 413},
				{"oversized", `{"email":"` + strings.Repeat("a", 1<<20) + `"}`, "Request Entity Too Large", "body must not be larger than 1048576 bytes", 413},
				{"missing required fields", `{}`, "Bad Request", "One or more fields are invalid. Check the field errors for details.", 400},
			} {
				t.Run(tt.name, func(t *testing.T) {
					t.Parallel()
					mux := http.NewServeMux()
					// Nil dependencies prove parsing/validation stops before business logic.
					auth.NewHandler(nil, slog.New(slog.NewTextHandler(io.Discard, nil)), false).Register(mux)
					w := httptest.NewRecorder()
					mux.ServeHTTP(w, httptest.NewRequest(http.MethodPost, path, strings.NewReader(tt.body)))
					assertHTTPError(t, w, tt.status, tt.title, tt.message)
				})
			}
		})
	}
}

func TestHandler_ValidationDetails(t *testing.T) {
	t.Parallel()
	fields := map[string][]string{
		"name":     {"This field is required"},
		"email":    {"This field is required"},
		"password": {"This field is required"},
	}
	for _, path := range []string{"/auth/register", "/auth/login"} {
		t.Run(path, func(t *testing.T) {
			t.Parallel()
			mux := http.NewServeMux()
			auth.NewHandler(nil, slog.New(slog.NewTextHandler(io.Discard, nil)), false).Register(mux)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{}`)))
			assertHTTPError(t, w, http.StatusBadRequest, "Bad Request", "One or more fields are invalid. Check the field errors for details.")
			var body struct{ Errors map[string][]string }
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
				t.Fatal(err)
			}
			want := fields
			if path == "/auth/login" {
				want = map[string][]string{"email": fields["email"], "password": fields["password"]}
			}
			if !reflect.DeepEqual(body.Errors, want) {
				t.Errorf("errors = %+v, want %+v", body.Errors, want)
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
			h := middleware.GetUserFromSession(func(context.Context, string) (string, error) { return "user-id", nil })(mux)
			r := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.payload))
			r.AddCookie(&http.Cookie{Name: "session", Value: "raw-token"})
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)
			assertHTTPError(t, w, http.StatusInternalServerError, "Internal Server Error", "internal error")
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
			assertHTTPError(t, w, http.StatusUnauthorized, "Unauthorized", "Session is expired or invalid")
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
	h := middleware.GetUserFromSession(func(context.Context, string) (string, error) { return "deleted-user", nil })(mux)
	r := httptest.NewRequest("GET", "/auth/me", nil)
	r.AddCookie(&http.Cookie{Name: "session", Value: "raw-token"})
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	assertHTTPError(t, w, http.StatusUnauthorized, "Unauthorized", "authentication required")
	if w.Header().Get("Cache-Control") != "no-store" || len(w.Result().Cookies()) != 0 {
		t.Error("unauthorized response must not be cached or change cookies")
	}
}

func assertHTTPError(t *testing.T, w *httptest.ResponseRecorder, status int, title, message string) {
	t.Helper()
	var body struct {
		Type, Title, Detail, Instance string
		Status                        int
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("invalid error JSON: %v", err)
	}
	if w.Code != status || body.Status != status || body.Type != "about:blank" || body.Title != title || body.Instance == "" || (message != "" && body.Detail != message) {
		t.Errorf("response = %d %s, want %d %s %q", w.Code, w.Body.String(), status, title, message)
	}
	if w.Header().Get("Content-Type") != "application/problem+json" {
		t.Error("error response must have JSON content type")
	}
}
