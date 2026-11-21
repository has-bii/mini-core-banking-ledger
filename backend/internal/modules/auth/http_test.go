package auth_test

import (
	"backend/internal/modules/auth"
	"backend/internal/modules/session"
	"backend/internal/modules/user"
	"backend/internal/platform/middleware"
	"backend/internal/testutil"
	"context"
	"encoding/json/v2"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

const testSessionLifetime = 7 * time.Minute

type authServiceStub struct {
	registerInput auth.RegisterDTO
	loginInput    auth.LoginDTO
	meUserID      string
	logoutTokens  []string
	account       user.User
	expiresAt     time.Time
	token         string
	err           error
}

func (s *authServiceStub) Register(_ context.Context, input auth.RegisterDTO) (user.User, error) {
	s.registerInput = input
	return s.account, s.err
}
func (s *authServiceStub) Login(_ context.Context, input auth.LoginDTO) (user.User, session.Session, string, error) {
	s.loginInput = input
	return s.account, session.Session{ExpiresAt: s.expiresAt}, s.token, s.err
}
func (s *authServiceStub) Me(_ context.Context, userID string) (user.User, error) {
	s.meUserID = userID
	return s.account, s.err
}
func (s *authServiceStub) Logout(_ context.Context, token string) error {
	s.logoutTokens = append(s.logoutTokens, token)
	return s.err
}

func TestHandler_ForwardsNormalizedDTOs(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, path, body string
		status           int
		wantRegister     auth.RegisterDTO
		wantLogin        auth.LoginDTO
	}{
		{
			name: "register", path: "/auth/register", status: http.StatusCreated,
			body:         `{"name":" Alice ","email":" alice@example.com ","password":" password "}`,
			wantRegister: auth.RegisterDTO{Name: "Alice", Email: "alice@example.com", Password: " password "},
		},
		{
			name: "login", path: "/auth/login", status: http.StatusOK,
			body:      `{"email":" alice@example.com ","password":" password "}`,
			wantLogin: auth.LoginDTO{Email: "alice@example.com", Password: " password "},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &authServiceStub{}
			mux := http.NewServeMux()
			auth.NewHandler(svc, nil, false, testSessionLifetime).Register(mux)
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, tt.path, strings.NewReader(tt.body)).WithContext(t.Context())
			mux.ServeHTTP(w, r)
			if w.Code != tt.status {
				t.Fatalf("status = %d, want %d", w.Code, tt.status)
			}
			if svc.registerInput != tt.wantRegister || svc.loginInput != tt.wantLogin {
				t.Error("service did not receive the expected normalized fields and unchanged password")
			}
		})
	}
}

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
					auth.NewHandler(nil, slog.New(slog.NewTextHandler(io.Discard, nil)), false, testSessionLifetime).Register(mux)
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
			auth.NewHandler(nil, slog.New(slog.NewTextHandler(io.Discard, nil)), false, testSessionLifetime).Register(mux)
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{}`)))
			assertHTTPError(t, w, http.StatusBadRequest, "Bad Request", "One or more fields are invalid. Check the field errors for details.")
			var body struct {
				Errors map[string][]string `json:"errors"`
			}
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
			auth.NewHandler(&authServiceStub{err: errors.New("secret database credentials")}, slog.New(slog.NewTextHandler(io.Discard, nil)), false, testSessionLifetime).Register(mux)
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
	auth.NewHandler(nil, slog.New(slog.NewTextHandler(io.Discard, nil)), false, testSessionLifetime).Register(mux)
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
	auth.NewHandler(&authServiceStub{err: auth.ErrUnauthenticated}, slog.New(slog.NewTextHandler(io.Discard, nil)), false, testSessionLifetime).Register(mux)
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

func TestHandler_DomainErrorResponses(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, method, path, payload, detail string
		err                                 error
		status                              int
	}{
		{"duplicate email", http.MethodPost, "/auth/register", `{"name":"Alice","email":"alice@example.com","password":"password123"}`, "email already taken", user.ErrEmailTaken, http.StatusConflict},
		{"invalid credentials", http.MethodPost, "/auth/login", `{"email":"alice@example.com","password":"incorrect"}`, "invalid email or password", auth.ErrInvalidCredentials, http.StatusUnauthorized},
		{"invalid resolved identity", http.MethodGet, "/auth/me", "", "internal error", user.ErrInvalidUserID, http.StatusInternalServerError},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mux := http.NewServeMux()
			auth.NewHandler(&authServiceStub{err: tt.err}, slog.New(slog.NewTextHandler(io.Discard, nil)), false, testSessionLifetime).Register(mux)
			ctx := t.Context()
			if tt.path == "/auth/me" {
				ctx = middleware.WithUserID(ctx, "not-a-uuid")
			}
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, httptest.NewRequestWithContext(ctx, tt.method, tt.path, strings.NewReader(tt.payload)))
			assertHTTPError(t, w, tt.status, http.StatusText(tt.status), tt.detail)
			if tt.path != "/auth/register" && w.Header().Get("Cache-Control") != "no-store" {
				t.Error("failed authentication response must not be cached")
			}
			if len(w.Result().Cookies()) != 0 {
				t.Error("failed request changed the session cookie")
			}
		})
	}
}

func TestHandler_UserResponses(t *testing.T) {
	t.Parallel()
	want := user.User{ID: "user-id", Name: "Alice", Email: "alice@example.com"}
	account := want
	account.Password = "private-password-hash"
	const token = "raw-session-token"
	expires := time.Date(2030, time.January, 2, 12, 0, 0, 0, time.UTC)
	for _, tt := range []struct {
		name, method, path, payload string
		status                      int
		secure                      bool
		lifetime                    time.Duration
		maxAge                      string
	}{
		{"register", http.MethodPost, "/auth/register", `{"name":"Alice","email":"alice@example.com","password":"password123"}`, http.StatusCreated, false, testSessionLifetime, ""},
		{"development login", http.MethodPost, "/auth/login", `{"email":"alice@example.com","password":"password123"}`, http.StatusOK, false, testSessionLifetime, "420"},
		{"production login with large Max-Age", http.MethodPost, "/auth/login", `{"email":"alice@example.com","password":"password123"}`, http.StatusOK, true, 1 << 31 * time.Second, "2147483648"},
		{"current user", http.MethodGet, "/auth/me", "", http.StatusOK, false, testSessionLifetime, ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &authServiceStub{account: account, token: token, expiresAt: expires}
			mux := http.NewServeMux()
			auth.NewHandler(svc, nil, tt.secure, tt.lifetime).Register(mux)
			ctx := t.Context()
			if tt.path == "/auth/me" {
				ctx = middleware.WithUserID(ctx, want.ID)
			}
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, httptest.NewRequestWithContext(ctx, tt.method, tt.path, strings.NewReader(tt.payload)))
			if w.Code != tt.status || w.Header().Get("Content-Type") != "application/json" {
				t.Fatalf("response = %d %v, want %d and JSON", w.Code, w.Header(), tt.status)
			}
			if tt.path == "/auth/me" && svc.meUserID != want.ID {
				t.Errorf("current user lookup used %q, want %q", svc.meUserID, want.ID)
			}
			var body struct {
				Data user.User `json:"data"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil || body.Data != want {
				t.Fatalf("unexpected public user response (decode error: %v)", err)
			}
			for _, secret := range []string{`"password"`, account.Password, `"token"`, `"session"`, token} {
				if strings.Contains(w.Body.String(), secret) {
					t.Error("public user response exposes a password or session field")
				}
			}
			if tt.path != "/auth/register" && w.Header().Get("Cache-Control") != "no-store" {
				t.Error("authentication response must not be cached")
			}
			cookies := w.Result().Cookies()
			if tt.path != "/auth/login" {
				if len(cookies) != 0 {
					t.Error("only login should issue a session cookie")
				}
				return
			}
			if len(cookies) != 1 {
				t.Fatalf("login sent %d cookies, want one", len(cookies))
			}
			c := cookies[0]
			if c.Name != "session" || c.Value != token || c.Path != "/" || !c.HttpOnly || c.Secure != tt.secure || c.SameSite != http.SameSiteLaxMode || !c.Expires.Equal(expires) {
				t.Error("login sent an incorrect session cookie")
			}
			if !strings.Contains(w.Header().Get("Set-Cookie")+";", "; Max-Age="+tt.maxAge+";") {
				t.Errorf("login cookie must have Max-Age=%s", tt.maxAge)
			}
		})
	}
}

func TestHandler_LogoutClearsSessionCookie(t *testing.T) {
	t.Parallel()
	const token = "raw-session-token"
	for _, tt := range []struct {
		name               string
		secure, withCookie bool
	}{
		{"development", false, true},
		{"production", true, true},
		{"without a session", true, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &authServiceStub{}
			mux := http.NewServeMux()
			auth.NewHandler(svc, nil, tt.secure, testSessionLifetime).Register(mux)
			r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/auth/logout", nil)
			var wantTokens []string
			if tt.withCookie {
				r.AddCookie(&http.Cookie{Name: "session", Value: token})
				wantTokens = []string{token}
			}
			w := httptest.NewRecorder()
			mux.ServeHTTP(w, r)
			if w.Code != http.StatusNoContent || w.Body.Len() != 0 || w.Header().Get("Cache-Control") != "no-store" {
				t.Errorf("logout response = %d %v %q, want empty 204 and no-store", w.Code, w.Header(), w.Body.String())
			}
			if !reflect.DeepEqual(svc.logoutTokens, wantTokens) {
				t.Error("logout must revoke exactly the presented session, if any")
			}
			cookies := w.Result().Cookies()
			if len(cookies) != 1 {
				t.Fatalf("logout sent %d cookies, want one", len(cookies))
			}
			c := cookies[0]
			if c.Name != "session" || c.Value != "" || c.MaxAge != -1 || c.Expires.IsZero() || !c.Expires.Before(time.Now()) || c.Path != "/" || !c.HttpOnly || c.Secure != tt.secure || c.SameSite != http.SameSiteLaxMode {
				t.Error("logout did not securely clear the session cookie")
			}
		})
	}
}

func assertHTTPError(t *testing.T, w *httptest.ResponseRecorder, status int, title, message string) {
	t.Helper()
	var body testutil.ProblemDetails
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
