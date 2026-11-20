//go:build integration

package app_test

import (
	"backend/internal/app"
	"backend/internal/config"
	"backend/internal/modules/user"
	"backend/internal/modules/wallet"
	"backend/internal/testutil"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestApp_AuthLifecycle(t *testing.T) {
	t.Parallel()
	pool := testutil.Postgres(t)
	redisClient := testutil.Redis(t)
	const lifetime = 45 * time.Minute
	h, err := app.New(config.Config{
		Env: "test", CORSOrigin: "http://localhost:3000", RatelimitCapacity: 50, RatelimitRefillRate: 10,
		SessionLifetime: lifetime,
	}, pool, redisClient,
		slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	// This test owns wiring and workflow, not validation, storage, or cookie matrices.
	registered := authRequest(t, h, "POST", "/api/auth/register", `{"name":"Alice","email":"alice@example.com","password":"password123"}`, nil, http.StatusCreated)
	u := publicUser(t, registered)
	beforeLogin := time.Now().Truncate(time.Second)
	login := authRequest(t, h, "POST", "/api/auth/login", `{"email":"alice@example.com","password":"password123"}`, nil, http.StatusOK)
	if got := publicUser(t, login); got != u {
		t.Fatalf("login user = %+v, want %+v", got, u)
	}
	cookies := login.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != "session" || cookies[0].Value == "" {
		t.Fatalf("login cookies = %v, want session token", cookies)
	}
	cookie := cookies[0]
	if cookie.MaxAge != 2700 || cookie.Expires.Before(beforeLogin.Add(lifetime)) || cookie.Expires.After(time.Now().Truncate(time.Second).Add(lifetime)) {
		t.Errorf("login cookie has Max-Age=%d and Expires=%v, want a lifetime of %v", cookie.MaxAge, cookie.Expires, lifetime)
	}
	me := authRequest(t, h, "GET", "/api/auth/me", "", cookie, http.StatusOK)
	if got := publicUser(t, me); got != u {
		t.Fatalf("current user = %+v, want %+v", got, u)
	}
	wallets := authRequest(t, h, "GET", "/api/wallets", "", cookie, http.StatusOK)
	var walletBody struct {
		Data []struct {
			ID         string `json:"id"`
			UserID     string `json:"user_id"`
			WalletType string `json:"wallet_type"`
		}
	}
	if err := json.Unmarshal(wallets.Body.Bytes(), &walletBody); err != nil || len(walletBody.Data) != 1 {
		t.Fatalf("wallets response = %s (%v), want one wallet", wallets.Body.String(), err)
	}
	if walletBody.Data[0].ID == "" || walletBody.Data[0].WalletType != "main" {
		t.Errorf("wallets response = %s, want main wallet with an ID", wallets.Body.String())
	}
	if wallets.Header().Get("Cache-Control") != "no-store" {
		t.Error("wallets response must not be cached")
	}
	mainID := walletBody.Data[0].ID
	created := authRequest(t, h, "POST", "/api/wallets", `{"name":" Savings ","wallet_type":" savings "}`, cookie, http.StatusCreated)
	var createdBody struct {
		Data wallet.Wallet `json:"data"`
	}
	if err := json.Unmarshal(created.Body.Bytes(), &createdBody); err != nil {
		t.Fatal(err)
	}
	savings := createdBody.Data
	if savings.ID == "" || savings.UserID != u.ID || savings.Name != "Savings" || savings.Type != wallet.Savings || savings.SortOrder != 1 || savings.Balance != 0 || savings.CreatedAt.IsZero() {
		t.Errorf("created wallet = %+v, unexpected fields", savings)
	}
	if created.Header().Get("Cache-Control") != "no-store" {
		t.Error("created wallet response must not be cached")
	}
	savingsID := savings.ID
	for _, tt := range []struct {
		name string
		path string
	}{
		{name: "ascending wallets", path: "/api/wallets"},
		{name: "ignore descending order", path: "/api/wallets?order=%20DESC%20"},
		{name: "ignore invalid order", path: "/api/wallets?order=invalid"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			response := authRequest(t, h, "GET", tt.path, "", cookie, http.StatusOK)
			if err := json.Unmarshal(response.Body.Bytes(), &walletBody); err != nil || len(walletBody.Data) != 2 {
				t.Fatalf("wallets response = %s (%v), want two wallets", response.Body.String(), err)
			}
			if walletBody.Data[0].ID != mainID || walletBody.Data[1].ID != savingsID {
				t.Errorf("wallets response = %s, want main then savings", response.Body.String())
			}
			for _, a := range walletBody.Data {
				if a.UserID != u.ID {
					t.Errorf("wallet user_id = %q, want %q", a.UserID, u.ID)
				}
			}
		})
	}
	authRequest(t, h, "POST", "/api/auth/logout", "", cookie, http.StatusNoContent)
	authRequest(t, h, "GET", "/api/auth/me", "", cookie, http.StatusUnauthorized)
	authRequest(t, h, "GET", "/api/wallets", "", cookie, http.StatusUnauthorized)
}

func TestApp_SessionLookupFailure(t *testing.T) {
	t.Parallel()
	pool := testutil.Postgres(t)
	redisClient := testutil.Redis(t)
	h, err := app.New(config.Config{
		Env: "test", CORSOrigin: "http://localhost:3000", RatelimitCapacity: 50, RatelimitRefillRate: 10,
	}, pool, redisClient,
		slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil {
		t.Fatal(err)
	}
	cookie := &http.Cookie{Name: "session", Value: strings.Repeat("A", 43)}
	authRequest(t, h, "GET", "/api/auth/me", "", cookie, http.StatusUnauthorized)
	if err := redisClient.Close(); err != nil {
		t.Fatal(err)
	}
	w := authRequest(t, h, "GET", "/api/auth/me", "", cookie, http.StatusInternalServerError)
	var body struct {
		Type, Title, Detail, Instance string
		Status                        int
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if w.Header().Get("Content-Type") != "application/problem+json" || body.Type != "about:blank" || body.Title != "Internal Server Error" || body.Status != 500 || body.Detail != "Unexpected error has occurred" || body.Instance != "/api/auth/me" {
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
