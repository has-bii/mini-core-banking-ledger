//go:build integration

package app_test

import (
	"backend/internal/app"
	"backend/internal/config"
	"backend/internal/modules/account"
	"backend/internal/modules/session"
	"backend/internal/modules/user"
	"backend/internal/testutil"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

func TestApp_AuthLifecycle(t *testing.T) {
	t.Parallel()
	pool := testutil.Postgres(t)
	h, err := app.New(config.Config{Env: "test", CORSOrigin: "http://localhost:3000"}, pool,
		slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	registration := `{"name":" Alice ","email":" alice@example.com ","password":"password123"}`
	registered := authRequest(t, h, "POST", "/api/auth/register", registration, nil, http.StatusCreated)
	u := publicUser(t, registered)
	if u.Name != "Alice" || u.Email != "alice@example.com" || u.CreatedAt.IsZero() || u.UpdatedAt.IsZero() {
		t.Fatalf("registered user = %+v", u)
	}
	if len(registered.Result().Cookies()) != 0 {
		t.Fatal("registration must not create a session cookie")
	}
	credentials, err := account.NewStore(pool).FindByUserID(t.Context(), u.ID)
	if err != nil {
		t.Fatal(err)
	}
	if credentials.Password == nil || *credentials.Password == "password123" {
		t.Fatal("password was not stored as a hash")
	}
	if err := bcrypt.CompareHashAndPassword([]byte(*credentials.Password), []byte("password123")); err != nil {
		t.Fatalf("stored password is not a valid bcrypt hash: %v", err)
	}

	duplicate := authRequest(t, h, "POST", "/api/auth/register", registration, nil, http.StatusConflict)
	var conflict struct{ Error struct{ Code string } }
	if err := json.Unmarshal(duplicate.Body.Bytes(), &conflict); err != nil || conflict.Error.Code != "EMAIL_TAKEN" {
		t.Fatalf("duplicate registration = %s (%v)", duplicate.Body.String(), err)
	}
	for _, tt := range []struct {
		name string
		body string
	}{
		{"unknown user", `{"email":"missing@example.com","password":"password123"}`},
		{"wrong password", `{"email":"alice@example.com","password":"incorrect"}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			w := authRequest(t, h, "POST", "/api/auth/login", tt.body, nil, http.StatusUnauthorized)
			var body struct{ Error struct{ Code string } }
			if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil || body.Error.Code != "INVALID_CREDENTIALS" || len(w.Result().Cookies()) != 0 {
				t.Fatalf("failed login = %s (%v), want invalid credentials without a cookie", w.Body.String(), err)
			}
		})
	}
	login := authRequest(t, h, "POST", "/api/auth/login", `{"email":"alice@example.com","password":"password123"}`, nil, http.StatusOK)
	if got := publicUser(t, login); got != u {
		t.Fatalf("login user = %+v, want %+v", got, u)
	}
	cookies := login.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "session" || cookies[0].Value == "" {
		t.Fatalf("login cookies = %v, want session token", cookies)
	}
	cookie := cookies[0]
	if strings.Contains(login.Body.String(), cookie.Value) {
		t.Fatal("raw session token leaked into JSON")
	}
	digest := sha256.Sum256([]byte(cookie.Value))
	store := session.NewStore(pool)
	stored, err := store.FindByTokenHash(t.Context(), hex.EncodeToString(digest[:]))
	if err != nil || stored.UserID != u.ID {
		t.Fatalf("session was not persisted under its SHA-256 hash: %+v, %v", stored, err)
	}
	if _, err := store.FindByTokenHash(t.Context(), cookie.Value); !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("raw token must not be persisted: %v", err)
	}

	me := authRequest(t, h, "GET", "/api/auth/me", "", cookie, http.StatusOK)
	if got := publicUser(t, me); got != u {
		t.Fatalf("current user = %+v, want %+v", got, u)
	}
	authRequest(t, h, "GET", "/api/auth/me", "", nil, http.StatusUnauthorized)
	authRequest(t, h, "GET", "/api/auth/me", "", &http.Cookie{Name: "session", Value: "malformed"}, http.StatusUnauthorized)
	logout := authRequest(t, h, "POST", "/api/auth/logout", "", cookie, http.StatusNoContent)
	cleared := logout.Result().Cookies()
	if len(cleared) != 1 || cleared[0].Value != "" || cleared[0].MaxAge != -1 {
		t.Fatal("logout must clear the session cookie")
	}
	authRequest(t, h, "GET", "/api/auth/me", "", cookie, http.StatusUnauthorized)
	if _, err := store.FindByTokenHash(t.Context(), hex.EncodeToString(digest[:])); !errors.Is(err, session.ErrNotFound) {
		t.Fatalf("logout did not delete persisted session: %v", err)
	}
	authRequest(t, h, "POST", "/api/auth/logout", "", cookie, http.StatusNoContent)
	authRequest(t, h, "POST", "/api/auth/logout", "", nil, http.StatusNoContent)
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
	for _, secret := range []string{"password", "token", "session"} {
		if strings.Contains(w.Body.String(), secret) {
			t.Fatalf("user response exposes %q", secret)
		}
	}
	return body.Data
}
