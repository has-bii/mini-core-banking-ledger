//go:build integration

package app_test

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"backend/internal/app"
	"backend/internal/config"
	"backend/internal/modules/user"
	"backend/internal/testutil/testdb"

	"github.com/google/uuid"
)

const (
	testEmail    = "alice@example.com"
	registration = `{"name":"Alice","email":"alice@example.com","password":"password123"}`
	credentials  = `{"email":"alice@example.com","password":"password123"}`
)

func newTestApp(t *testing.T, env string) http.Handler {
	t.Helper()
	handler, err := app.New(config.Config{
		Env: env, CORSOrigin: "http://localhost:3000", RateLimitRPS: 100, RateLimitBurst: 200,
	}, testdb.Open(t), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	return handler
}

func request(t *testing.T, handler http.Handler, method, path, body string, status int, cookie *http.Cookie) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, strings.NewReader(body)).WithContext(t.Context())
	if body != "" {
		req.Header.Set("Content-Type", "application/json")
	}
	if cookie != nil {
		req.AddCookie(cookie)
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	if rec.Code != status {
		t.Fatalf("%s %s: status = %d, want %d; body = %s", method, path, rec.Code, status, rec.Body.String())
	}
	return rec
}

func decodeUser(t *testing.T, rec *httptest.ResponseRecorder) user.User {
	t.Helper()
	var response struct {
		Data user.User `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode user: %v", err)
	}
	for _, secret := range []string{"password", "token"} {
		if strings.Contains(rec.Body.String(), secret) {
			t.Errorf("response exposed %s", secret)
		}
	}
	return response.Data
}

func sessionCookie(t *testing.T, rec *httptest.ResponseRecorder) *http.Cookie {
	t.Helper()
	for _, cookie := range rec.Result().Cookies() {
		if cookie.Name == "session" && cookie.Value != "" {
			return cookie
		}
	}
	t.Fatal("login did not set a session cookie")
	return nil
}

func register(t *testing.T, handler http.Handler) user.User {
	t.Helper()
	return decodeUser(t, request(t, handler, "POST", "/api/auth/register", registration, http.StatusCreated, nil))
}

func login(t *testing.T, handler http.Handler) *http.Cookie {
	t.Helper()
	return sessionCookie(t, request(t, handler, "POST", "/api/auth/login", credentials, http.StatusOK, nil))
}

func assertErrorCode(t *testing.T, rec *httptest.ResponseRecorder, code string) {
	t.Helper()
	var response struct {
		Error struct{ Code string } `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
		t.Fatal(err)
	}
	if response.Error.Code != code {
		t.Errorf("error code = %q, want %q", response.Error.Code, code)
	}
}

func TestAuthJourney(t *testing.T) {
	t.Parallel()
	handler := newTestApp(t, "test")
	registered := register(t, handler)
	if _, err := uuid.Parse(registered.ID); err != nil || registered.Email != testEmail || registered.Name != "Alice" {
		t.Fatalf("unexpected registered user: %+v", registered)
	}
	response := request(t, handler, "POST", "/api/auth/login", credentials, http.StatusOK, nil)
	if loggedIn := decodeUser(t, response); loggedIn != registered {
		t.Fatalf("login returned a different user: %+v", loggedIn)
	}
	cookie := sessionCookie(t, response)
	me := request(t, handler, "GET", "/api/auth/me", "", http.StatusOK, cookie)
	if current := decodeUser(t, me); current != registered {
		t.Fatalf("current user differs from registered user: %+v", current)
	}
	logout := request(t, handler, "POST", "/api/auth/logout", "", http.StatusNoContent, cookie)
	cookies := logout.Result().Cookies()
	if logout.Body.Len() != 0 || len(cookies) != 1 || cookies[0].Name != "session" || cookies[0].Value != "" || cookies[0].MaxAge != -1 {
		t.Fatalf("logout did not clear session cookie: body=%s; cookies=%v", logout.Body.String(), cookies)
	}
	rejected := request(t, handler, "GET", "/api/auth/me", "", http.StatusUnauthorized, cookie)
	assertErrorCode(t, rejected, "UNAUTHENTICATED")
}

func TestAuthErrorsIntegration(t *testing.T) {
	for _, tc := range []struct {
		name, method, path, body, code string
		status                         int
		registered                     bool
	}{
		{"duplicate registration", "POST", "/api/auth/register", registration, "EMAIL_TAKEN", 409, true},
		{"wrong password", "POST", "/api/auth/login", `{"email":"alice@example.com","password":"wrong-password"}`, "INVALID_CREDENTIALS", 401, true},
		{"unknown user", "POST", "/api/auth/login", credentials, "INVALID_CREDENTIALS", 401, false},
		{"unauthenticated", "GET", "/api/auth/me", "", "UNAUTHENTICATED", 401, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			handler := newTestApp(t, "test")
			if tc.registered {
				register(t, handler)
			}
			rec := request(t, handler, tc.method, tc.path, tc.body, tc.status, nil)
			assertErrorCode(t, rec, tc.code)
		})
	}
}

func TestLogoutOnlyRevokesCurrentSession(t *testing.T) {
	t.Parallel()
	handler := newTestApp(t, "test")
	register(t, handler)
	first, second := login(t, handler), login(t, handler)
	if first.Value == second.Value {
		t.Fatal("separate logins must create different sessions")
	}
	request(t, handler, "POST", "/api/auth/logout", "", http.StatusNoContent, first)
	request(t, handler, "GET", "/api/auth/me", "", http.StatusUnauthorized, first)
	request(t, handler, "GET", "/api/auth/me", "", http.StatusOK, second)
	request(t, handler, "POST", "/api/auth/logout", "", http.StatusNoContent, first)
}

func TestLogoutWithoutValidSession(t *testing.T) {
	t.Parallel()
	handler := newTestApp(t, "test")
	for _, tc := range []struct {
		name   string
		cookie *http.Cookie
	}{
		{"missing", nil},
		{"empty", &http.Cookie{Name: "session", Value: ""}},
		{"invalid", &http.Cookie{Name: "session", Value: "invalid"}},
		{"unknown", &http.Cookie{Name: "session", Value: strings.Repeat("A", 43)}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			request(t, handler, "POST", "/api/auth/logout", "", http.StatusNoContent, tc.cookie)
		})
	}
}

func TestSessionCookieEnvironment(t *testing.T) {
	for _, env := range []string{"dev", "test", "production"} {
		t.Run(env, func(t *testing.T) {
			t.Parallel()
			handler := newTestApp(t, env)
			register(t, handler)
			cookie := login(t, handler)
			if !cookie.HttpOnly || cookie.SameSite != http.SameSiteLaxMode || cookie.Path != "/" || cookie.Domain != "" || cookie.Secure != (env == "production") {
				t.Fatalf("unexpected %s session cookie: %+v", env, cookie)
			}
		})
	}
}
