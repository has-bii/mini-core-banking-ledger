//go:build integration

package app_test

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"backend/internal/app"
	"backend/internal/modules/user"
	"backend/internal/platform/database"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func TestAuthIntegration(t *testing.T) {
	// Read the file directly so a shell DATABASE_URL cannot override the test database.
	env, err := godotenv.Read(filepath.Join("..", "..", ".env.test"))
	if err != nil {
		t.Fatalf("read backend/.env.test: %v", err)
	}
	if env["DATABASE_URL"] == "" {
		t.Fatal("DATABASE_URL is required in backend/.env.test")
	}

	ctx, cancel := context.WithTimeout(t.Context(), 30*time.Second)
	defer cancel()
	pool, err := database.Open(ctx, env["DATABASE_URL"])
	if err != nil {
		t.Fatal("open test database (apply migrations first): ", err)
	}
	t.Cleanup(pool.Close)

	email := "integration-" + uuid.NewString() + "@example.com"
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := pool.Exec(cleanupCtx, `DELETE FROM users WHERE email = $1`, email); err != nil {
			t.Errorf("clean up test user: %v", err)
		}
	})

	handler := app.New("test", pool, slog.New(slog.NewTextHandler(io.Discard, nil)))
	request := func(t *testing.T, method, path, body string, status int, cookie *http.Cookie) *httptest.ResponseRecorder {
		t.Helper()
		req := httptest.NewRequest(method, path, strings.NewReader(body)).WithContext(ctx)
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
	decodeUser := func(rec *httptest.ResponseRecorder) user.User {
		t.Helper()
		var response struct {
			Data user.User `json:"data"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
			t.Fatalf("decode user: %v", err)
		}
		return response.Data
	}

	registration := `{"name":"Integration User","email":"` + email + `","password":"password123"}`
	registered := decodeUser(request(t, http.MethodPost, "/api/auth/register", registration, http.StatusCreated, nil))
	if _, err := uuid.Parse(registered.ID); err != nil || registered.Email != email || registered.Name != "Integration User" {
		t.Fatalf("unexpected registered user: %+v", registered)
	}

	for _, tc := range []struct {
		name, method, path, body, code string
		status                         int
	}{
		{"duplicate registration", http.MethodPost, "/api/auth/register", registration, "EMAIL_TAKEN", http.StatusConflict},
		{"wrong password", http.MethodPost, "/api/auth/login", `{"email":"` + email + `","password":"wrong-password"}`, "INVALID_CREDENTIALS", http.StatusUnauthorized},
		{"unauthenticated", http.MethodGet, "/api/auth/me", "", "UNAUTHENTICATED", http.StatusUnauthorized},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec := request(t, tc.method, tc.path, tc.body, tc.status, nil)
			var response struct {
				Error struct {
					Code string `json:"code"`
				} `json:"error"`
			}
			if err := json.Unmarshal(rec.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if response.Error.Code != tc.code {
				t.Errorf("error code = %q, want %q", response.Error.Code, tc.code)
			}
		})
	}

	login := request(t, http.MethodPost, "/api/auth/login", `{"email":"`+email+`","password":"password123"}`, http.StatusOK, nil)
	if loggedIn := decodeUser(login); loggedIn.ID != registered.ID || loggedIn.Email != email {
		t.Fatalf("login returned a different user: %+v", loggedIn)
	}
	var sessionCookie *http.Cookie
	for _, cookie := range login.Result().Cookies() {
		if cookie.Name == "session" {
			sessionCookie = cookie
		}
	}
	if sessionCookie == nil || sessionCookie.Value == "" {
		t.Fatal("login did not set a session cookie")
	}
	if !sessionCookie.HttpOnly || sessionCookie.SameSite != http.SameSiteLaxMode || sessionCookie.Path != "/" {
		t.Fatalf("unexpected session cookie attributes: %+v", sessionCookie)
	}

	otherLogin := request(t, http.MethodPost, "/api/auth/login", `{"email":"`+email+`","password":"password123"}`, http.StatusOK, nil)
	var otherCookie *http.Cookie
	for _, cookie := range otherLogin.Result().Cookies() {
		if cookie.Name == "session" {
			otherCookie = cookie
		}
	}
	if otherCookie == nil || otherCookie.Value == sessionCookie.Value {
		t.Fatal("second login did not create a separate session")
	}

	me := request(t, http.MethodGet, "/api/auth/me", "", http.StatusOK, sessionCookie)
	if current := decodeUser(me); current.ID != registered.ID || current.Email != email {
		t.Fatalf("current user differs from registered user: %+v", current)
	}

	logout := request(t, http.MethodPost, "/api/auth/logout", "", http.StatusNoContent, sessionCookie)
	cookies := logout.Result().Cookies()
	if logout.Body.Len() != 0 || len(cookies) != 1 || cookies[0].Name != "session" || cookies[0].Value != "" || cookies[0].MaxAge != -1 {
		t.Fatalf("logout did not clear session cookie: body=%s; cookies=%v", logout.Body.String(), cookies)
	}
	request(t, http.MethodGet, "/api/auth/me", "", http.StatusUnauthorized, sessionCookie)
	request(t, http.MethodGet, "/api/auth/me", "", http.StatusOK, otherCookie)
	request(t, http.MethodPost, "/api/auth/logout", "", http.StatusNoContent, sessionCookie)
	request(t, http.MethodPost, "/api/auth/logout", "", http.StatusNoContent, nil)
	request(t, http.MethodPost, "/api/auth/logout", "", http.StatusNoContent, &http.Cookie{Name: "session", Value: "invalid"})
}
