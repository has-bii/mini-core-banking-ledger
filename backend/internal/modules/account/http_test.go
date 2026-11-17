package account_test

import (
	"backend/internal/modules/account"
	"backend/internal/platform/transport/httpx/middleware"
	"backend/internal/platform/v2/httpx"
	v2Middleware "backend/internal/platform/v2/middleware"
	"backend/internal/platform/v2/validator"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

const accountTestUserID = "01a1c6de-805e-7b0c-a1b1-b851a9c38ccd"

func accountTestHandler(t *testing.T, store *accountStoreStub) http.Handler {
	t.Helper()
	mux := http.NewServeMux()
	account.NewHandler(account.NewService(store), slog.New(slog.NewTextHandler(io.Discard, nil))).Register(mux)
	return middleware.GetUser(func(context.Context, string) (string, error) {
		return accountTestUserID, nil
	})(mux)
}

func assertAccountProblem(t *testing.T, w *httptest.ResponseRecorder, status int, detail, instance string, fields validator.ErrorField) {
	t.Helper()
	if w.Code != status || w.Header().Get("Content-Type") != "application/problem+json" || w.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("response = %d %v %s", w.Code, w.Header(), w.Body.String())
	}
	var got httpx.ErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	var gotFields validator.ErrorField
	if got.Errors != nil {
		gotFields = *got.Errors
	}
	if got.Type != "about:blank" || got.Title != http.StatusText(status) || got.Status != status || got.Detail != detail || got.Instance != instance || !reflect.DeepEqual(gotFields, fields) {
		t.Errorf("unexpected problem details: %s; want %d %q, fields %+v", w.Body.String(), status, detail, fields)
	}
}

func TestHandler_PreservesV2Session(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name         string
		legacyCookie bool
	}{
		{name: "native v2 context"},
		{name: "v2 takes precedence over legacy", legacyCookie: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Force a storage failure to observe identity forwarding without SQL.
			store := &accountStoreStub{err: errors.New("database unavailable")}
			type requestKey struct{}
			ctx := context.WithValue(t.Context(), requestKey{}, "request-value")
			ctx = v2Middleware.WithUserID(ctx, "native-user")
			r := httptest.NewRequest(http.MethodGet, "/accounts", nil).WithContext(ctx)
			if tt.legacyCookie {
				r.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
			}
			w := httptest.NewRecorder()
			accountTestHandler(t, store).ServeHTTP(w, r)
			assertAccountProblem(t, w, http.StatusInternalServerError, "internal error", "/accounts", nil)
			if store.userID != "native-user" || store.ctx == nil || store.ctx.Value(requestKey{}) != "request-value" {
				t.Fatalf("identity or request context lost: %+v", store)
			}
		})
	}
}

func TestHandler_CreateAccountRejectsInvalidPayload(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, body, detail string
		status             int
		fields             validator.ErrorField
	}{
		{name: "malformed JSON", body: `{`, status: 400, detail: "body contains badly-formed JSON"},
		{name: "wrong field type", body: `{"name":123,"account_type":"savings"}`, status: 400, detail: `body contains an invalid value for field "name" (expected string)`},
		{name: "client supplied owner", body: `{"name":"Travel","account_type":"savings","user_id":"another-user"}`, status: 400, detail: `body contains unknown field "user_id"`},
		{name: "client supplied balance", body: `{"name":"Travel","account_type":"savings","balance":100}`, status: 400, detail: `body contains unknown field "balance"`},
		{name: "oversized body", body: `{"name":"` + strings.Repeat("A", 1<<20) + `","account_type":"savings"}`, status: 413, detail: "body must not be larger than 1048576 bytes"},
		{name: "missing fields", body: `{}`, status: 400, detail: "One or more fields are invalid. Check the field errors for details.", fields: validator.ErrorField{
			"name": {"This field is required"}, "account_type": {"This field is required"},
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := httptest.NewRequest(http.MethodPost, "/accounts", strings.NewReader(tt.body)).WithContext(t.Context())
			r.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
			w := httptest.NewRecorder()
			accountTestHandler(t, nil).ServeHTTP(w, r)
			assertAccountProblem(t, w, tt.status, tt.detail, "/accounts", tt.fields)
		})
	}
}

func TestHandler_CreateAccountErrors(t *testing.T) {
	t.Parallel()
	fields := validator.ErrorField{"name": {"This field is required"}}
	for _, tt := range []struct {
		name            string
		unauthenticated bool
		storeErr        error
		status          int
		detail          string
		fields          validator.ErrorField
	}{
		{name: "unauthenticated", unauthenticated: true, status: 401, detail: "Session is expired or invalid"},
		{name: "missing owner", storeErr: fmt.Errorf("create: %w", account.ErrNotFound), status: 404, detail: "account not found"},
		{name: "invalid user ID", storeErr: fmt.Errorf("create: %w", account.ErrInvalidUserID), status: 400, detail: "invalid user ID"},
		{name: "defensive validation", storeErr: fmt.Errorf("create: %w", &validator.ErrorValidation{Errors: &fields}), status: 400, detail: "One or more fields are invalid. Check the field errors for details.", fields: fields},
		{name: "store error", storeErr: errors.New("database password must not leak"), status: 500, detail: "internal error"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := &accountStoreStub{err: tt.storeErr}
			r := httptest.NewRequest(http.MethodPost, "/accounts", strings.NewReader(`{"name":"Travel","account_type":"savings"}`)).WithContext(t.Context())
			if !tt.unauthenticated {
				r.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
			}
			w := httptest.NewRecorder()
			accountTestHandler(t, store).ServeHTTP(w, r)
			assertAccountProblem(t, w, tt.status, tt.detail, "/accounts", tt.fields)
			if tt.unauthenticated && store.ctx != nil {
				t.Error("unauthenticated request reached store")
			}
		})
	}
}

func TestHandler_GetAccount(t *testing.T) {
	t.Parallel()
	const id = "01a1c6de-805e-7b0c-a1b1-b851a9c38cce"
	want := account.Account{ID: id, UserID: accountTestUserID, Name: "Main Account", Type: account.Main, Balance: 12345}
	for _, tt := range []struct {
		name, id, detail string
		unauthenticated  bool
		storeErr         error
		status           int
	}{
		{name: "returns account", id: id, status: 200},
		{name: "invalid account ID", id: "invalid", storeErr: fmt.Errorf("lookup: %w", account.ErrInvalidAccountID), status: 400, detail: "invalid account ID"},
		{name: "invalid user ID", id: id, storeErr: fmt.Errorf("lookup: %w", account.ErrInvalidUserID), status: 400, detail: "invalid user ID"},
		{name: "missing or unowned account", id: id, storeErr: fmt.Errorf("lookup: %w", account.ErrNotFound), status: 404, detail: "account not found"},
		{name: "store error", id: id, storeErr: errors.New("database connection timeout"), status: 500, detail: "internal error"},
		{name: "unauthenticated", id: id, unauthenticated: true, status: 401, detail: "Session is expired or invalid"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := &accountStoreStub{account: want, err: tt.storeErr}
			r := httptest.NewRequest(http.MethodGet, "/accounts/"+tt.id, nil).WithContext(t.Context())
			if !tt.unauthenticated {
				r.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
			}
			w := httptest.NewRecorder()
			accountTestHandler(t, store).ServeHTTP(w, r)
			if tt.unauthenticated {
				if store.ctx != nil {
					t.Error("unauthenticated request reached store")
				}
			} else if store.id != tt.id || store.userID != accountTestUserID {
				t.Error("store did not receive requested account ID and authenticated user ID")
			}
			if tt.status != http.StatusOK {
				assertAccountProblem(t, w, tt.status, tt.detail, r.URL.Path, nil)
				return
			}
			var got struct {
				Data account.Account `json:"data"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if w.Code != 200 || w.Header().Get("Content-Type") != "application/json" || w.Header().Get("Cache-Control") != "no-store" || got.Data != want {
				t.Fatalf("response = %d %s, want %+v", w.Code, w.Body.String(), want)
			}
		})
	}
}

func TestHandler_GetAccounts(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name     string
		accounts []account.Account
		storeErr error
		status   int
		detail   string
	}{
		{name: "returns accounts", accounts: []account.Account{{ID: "01a1c6de-805e-7b0c-a1b1-b851a9c38cce", UserID: accountTestUserID, Name: "Main Account", Type: account.Main, Balance: 12345}}, status: 200},
		{name: "empty list", accounts: []account.Account{}, status: 200},
		{name: "invalid user ID", storeErr: account.ErrInvalidUserID, status: 400, detail: "invalid user ID"},
		{name: "store error", storeErr: errors.New("database connection timeout"), status: 500, detail: "internal error"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := &accountStoreStub{accounts: tt.accounts, err: tt.storeErr}
			r := httptest.NewRequest(http.MethodGet, "/accounts", nil).WithContext(t.Context())
			r.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
			w := httptest.NewRecorder()
			accountTestHandler(t, store).ServeHTTP(w, r)
			if store.userID != accountTestUserID {
				t.Error("store did not receive authenticated user ID")
			}
			if tt.status != http.StatusOK {
				assertAccountProblem(t, w, tt.status, tt.detail, "/accounts", nil)
				return
			}
			var got struct {
				Data []account.Account `json:"data"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if w.Code != 200 || w.Header().Get("Content-Type") != "application/json" || w.Header().Get("Cache-Control") != "no-store" || !reflect.DeepEqual(got.Data, tt.accounts) {
				t.Fatalf("response = %d %s, want %+v", w.Code, w.Body.String(), tt.accounts)
			}
		})
	}
}
