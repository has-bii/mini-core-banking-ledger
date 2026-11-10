//go:build integration

package app_test

import (
	"backend/internal/app"
	"backend/internal/config"
	"backend/internal/modules/user"
	"backend/internal/testutil"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestApp_AuthLifecycle(t *testing.T) {
	t.Parallel()
	pool := testutil.Postgres(t)
	h, err := app.New(config.Config{
		Env: "test", CORSOrigin: "http://localhost:3000", RatelimitCapacity: 50, RatelimitRefillRate: 10,
	}, pool,
		slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	// This test owns wiring and workflow, not validation, storage, or cookie matrices.
	registered := authRequest(t, h, "POST", "/api/auth/register", `{"name":"Alice","email":"alice@example.com","password":"password123"}`, nil, http.StatusCreated)
	u := publicUser(t, registered)
	login := authRequest(t, h, "POST", "/api/auth/login", `{"email":"alice@example.com","password":"password123"}`, nil, http.StatusOK)
	if got := publicUser(t, login); got != u {
		t.Fatalf("login user = %+v, want %+v", got, u)
	}
	cookies := login.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "session" || cookies[0].Value == "" {
		t.Fatalf("login cookies = %v, want session token", cookies)
	}
	cookie := cookies[0]
	me := authRequest(t, h, "GET", "/api/auth/me", "", cookie, http.StatusOK)
	if got := publicUser(t, me); got != u {
		t.Fatalf("current user = %+v, want %+v", got, u)
	}
	authRequest(t, h, "POST", "/api/auth/logout", "", cookie, http.StatusNoContent)
	authRequest(t, h, "GET", "/api/auth/me", "", cookie, http.StatusUnauthorized)
}

func TestApp_SessionLookupFailure(t *testing.T) {
	t.Parallel()
	pool := testutil.Postgres(t)
	h, err := app.New(config.Config{
		Env: "test", CORSOrigin: "http://localhost:3000", RatelimitCapacity: 50, RatelimitRefillRate: 10,
	}, pool,
		slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	cookie := &http.Cookie{Name: "session", Value: strings.Repeat("A", 43)}
	authRequest(t, h, "GET", "/api/auth/me", "", cookie, http.StatusUnauthorized)
	pool.Close()
	w := authRequest(t, h, "GET", "/api/auth/me", "", cookie, http.StatusInternalServerError)
	var body struct {
		Error struct{ Code, Message string }
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body.Error.Code != "INTERNAL_ERROR" || body.Error.Message != "internal error" {
		t.Fatalf("lookup failure = %s, want a generic internal error", w.Body.String())
	}
}

func authRequest(t *testing.T, h http.Handler, method, path, body string, cookie *http.Cookie, status int) *httptest.ResponseRecorder {
	t.Helper()
	r := httptest.NewRequest(method, path, strings.NewReader(body)).WithContext(t.Context())
	r.Header.Set("Content-Type", "application/json")
	if cookie != nil {
		r.AddCookie(cookie)
	}
	w := httptest.NewRecorder()
	h.ServeHTTP(w, r)
	if w.Code != status {
		t.Fatalf("%s %s: status=%d body=%s, want %d", method, path, w.Code, w.Body.String(), status)
	}
	return w
}

func publicUser(t *testing.T, w *httptest.ResponseRecorder) user.User {
	t.Helper()
	var body struct{ Data user.User }
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil || body.Data.ID == "" {
		t.Fatalf("invalid user response: %s (%v)", w.Body.String(), err)
	}
	return body.Data
}
