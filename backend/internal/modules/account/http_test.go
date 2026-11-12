package account_test

import (
	"backend/internal/modules/account"
	"backend/internal/platform/transport/httpx"
	"backend/internal/platform/transport/httpx/middleware"
	"backend/internal/platform/validation"
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

func TestHandler_CreateAccountRejectsInvalidPayload(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, body, code, message string
		status                    int
		details                   validation.Errors
	}{
		{name: "malformed JSON", body: `{`, status: 400, code: "INVALID_JSON", message: "Request body contains malformed JSON."},
		{name: "wrong field type", body: `{"name":123,"account_type":"savings"}`, status: 400, code: "INVALID_JSON", message: `Field "name" must be a string.`, details: validation.Errors{
			{Field: "name", Code: "INVALID_TYPE", Message: `Field "name" must be a string.`},
		}},
		{name: "client supplied owner", body: `{"name":"Travel","account_type":"savings","user_id":"another-user"}`, status: 400, code: "INVALID_JSON", message: `Field "user_id" is not allowed.`, details: validation.Errors{
			{Field: "user_id", Code: "UNKNOWN_FIELD", Message: `Field "user_id" is not allowed.`},
		}},
		{name: "client supplied balance", body: `{"name":"Travel","account_type":"savings","balance":100}`, status: 400, code: "INVALID_JSON", message: `Field "balance" is not allowed.`, details: validation.Errors{
			{Field: "balance", Code: "UNKNOWN_FIELD", Message: `Field "balance" is not allowed.`},
		}},
		{name: "oversized body", body: `{"name":"` + strings.Repeat("A", 1<<20) + `","account_type":"savings"}`, status: 413, code: "PAYLOAD_TOO_LARGE", message: "Request body must not exceed 1 MiB."},
		{name: "missing fields", body: `{}`, status: 400, code: "VALIDATION_ERROR", message: "Some fields are invalid.", details: validation.Errors{
			{Field: "name", Code: "REQUIRED", Message: "Name is required."},
			{Field: "account_type", Code: "REQUIRED", Message: "Account type is required."},
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			mux := http.NewServeMux()
			account.NewHandler(nil, nil).Register(mux)
			handler := middleware.GetUser(func(context.Context, string) (string, error) {
				return "01a1c6de-805e-7b0c-a1b1-b851a9c38ccd", nil
			})(mux)
			r := httptest.NewRequest(http.MethodPost, "/accounts", strings.NewReader(tt.body)).WithContext(t.Context())
			r.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			if w.Code != tt.status {
				t.Fatalf("status = %d, want %d; body = %s", w.Code, tt.status, w.Body.String())
			}
			if got := w.Header().Get("Cache-Control"); got != "no-store" {
				t.Errorf("Cache-Control = %q, want no-store", got)
			}
			var got struct {
				Error struct {
					Code, Message string
					Details       validation.Errors
				}
			}
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if got.Error.Code != tt.code || got.Error.Message != tt.message || !reflect.DeepEqual(got.Error.Details, tt.details) {
				t.Errorf("response = %s, want %s %q with details %+v", w.Body.String(), tt.code, tt.message, tt.details)
			}
		})
	}
}

func TestHandler_CreateAccountErrors(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name            string
		unauthenticated bool
		storeErr        error
		status          int
		code            string
		message         string
	}{
		{name: "unauthenticated", unauthenticated: true, status: http.StatusUnauthorized, code: httpx.CodeUnauthenticated, message: "authentication required"},
		{name: "store error", storeErr: errors.New("database password must not leak"), status: http.StatusInternalServerError, code: httpx.CodeInternalError, message: "internal server error"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := &accountStoreStub{err: tt.storeErr}
			mux := http.NewServeMux()
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			account.NewHandler(account.NewService(store), logger).Register(mux)
			handler := middleware.GetUser(func(context.Context, string) (string, error) {
				return "01a1c6de-805e-7b0c-a1b1-b851a9c38ccd", nil
			})(mux)
			r := httptest.NewRequest(http.MethodPost, "/accounts", strings.NewReader(`{"name":"Travel","account_type":"savings"}`)).WithContext(t.Context())
			if !tt.unauthenticated {
				r.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)
			if w.Code != tt.status {
				t.Fatalf("status = %d, want %d; body = %s", w.Code, tt.status, w.Body.String())
			}
			if tt.unauthenticated && store.ctx != nil {
				t.Error("unauthenticated request reached store")
			}
			var got map[string]map[string]string
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			want := map[string]map[string]string{"error": {"code": tt.code, "message": tt.message}}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("response = %v, want %v", got, want)
			}
		})
	}
}

func TestHandler_GetAccount(t *testing.T) {
	t.Parallel()
	const userID = "01a1c6de-805e-7b0c-a1b1-b851a9c38ccd"
	const id = "01a1c6de-805e-7b0c-a1b1-b851a9c38cce"
	want := account.Account{ID: id, UserID: userID, Name: "Main Account", Type: account.Main, Balance: 12345}

	for _, tt := range []struct {
		name            string
		id              string
		unauthenticated bool
		serviceErr      error
		status          int
		code            string
		message         string
	}{
		{name: "returns account", id: id, status: http.StatusOK},
		{name: "invalid account ID", id: "invalid", serviceErr: fmt.Errorf("account by id: ID %w", validation.ErrInvalidUUID), status: http.StatusBadRequest, code: httpx.CodeValidationError, message: "invalid account ID"},
		{name: "missing or unowned account", id: id, serviceErr: fmt.Errorf("lookup: %w", account.ErrNotFound), status: http.StatusNotFound, code: httpx.CodeNotFound, message: "account not found"},
		{name: "store error", id: id, serviceErr: errors.New("database connection timeout"), status: http.StatusInternalServerError, code: httpx.CodeInternalError, message: "internal server error"},
		{name: "unauthenticated", id: id, unauthenticated: true, status: http.StatusUnauthorized, code: httpx.CodeUnauthenticated, message: "authentication required"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := &accountStoreStub{account: want, err: tt.serviceErr}
			mux := http.NewServeMux()
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			account.NewHandler(account.NewService(store), logger).Register(mux)

			var requestCtx context.Context
			handler := middleware.GetUser(func(context.Context, string) (string, error) {
				return userID, nil
			})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestCtx = r.Context()
				mux.ServeHTTP(w, r)
			}))
			r := httptest.NewRequest(http.MethodGet, "/accounts/"+tt.id, nil).WithContext(t.Context())
			if !tt.unauthenticated {
				r.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, r)

			if w.Code != tt.status {
				t.Fatalf("status = %d, want %d; body = %s", w.Code, tt.status, w.Body.String())
			}
			if tt.unauthenticated {
				if store.ctx != nil {
					t.Error("unauthenticated request reached store")
				}
			} else {
				if store.id != tt.id || store.userID != userID || store.ctx != requestCtx {
					t.Error("store did not receive account ID, authenticated user ID and request context")
				}
				if got := w.Header().Get("Cache-Control"); got != "no-store" {
					t.Errorf("Cache-Control = %q, want no-store", got)
				}
			}

			if tt.status != http.StatusOK {
				var got map[string]map[string]string
				if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
					t.Fatal(err)
				}
				want := map[string]map[string]string{"error": {"code": tt.code, "message": tt.message}}
				if !reflect.DeepEqual(got, want) {
					t.Errorf("response = %v, want %v", got, want)
				}
				return
			}

			var got struct {
				Data account.Account `json:"data"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got.Data, want) {
				t.Errorf("account = %+v, want %+v", got.Data, want)
			}
		})
	}
}

func TestHandler_GetAccounts(t *testing.T) {
	t.Parallel()
	const userID = "01a1c6de-805e-7b0c-a1b1-b851a9c38ccd"

	for _, tt := range []struct {
		name           string
		accounts       []account.Account
		serviceErr     error
		expectedStatus int
	}{
		{
			name: "Happy Path: returns accounts successfully (200 OK)",
			accounts: []account.Account{
				{ID: "01a1c6de-805e-7b0c-a1b1-b851a9c38cce", UserID: userID, Name: "Main Account", Type: account.Main, Balance: 12345},
			},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Happy Path: returns empty list when user has no accounts (200 OK)",
			accounts:       []account.Account{},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Error Path: returns 500 Internal Server Error when service fails",
			serviceErr:     errors.New("database connection timeout"),
			expectedStatus: http.StatusInternalServerError,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			store := &accountStoreStub{accounts: tt.accounts, err: tt.serviceErr}
			mux := http.NewServeMux()
			logger := slog.New(slog.NewTextHandler(io.Discard, nil))
			account.NewHandler(account.NewService(store), logger).Register(mux)

			var requestCtx context.Context
			testHandler := middleware.GetUser(func(context.Context, string) (string, error) {
				return userID, nil
			})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestCtx = r.Context()
				mux.ServeHTTP(w, r)
			}))
			r := httptest.NewRequest(http.MethodGet, "/accounts", nil).WithContext(t.Context())
			r.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
			w := httptest.NewRecorder()

			testHandler.ServeHTTP(w, r)

			if store.userID != userID {
				t.Errorf("service received userID = %q, want %q", store.userID, userID)
			}
			if store.ctx != requestCtx {
				t.Error("service did not receive the authenticated request context")
			}

			if w.Code != tt.expectedStatus {
				t.Fatalf("status = %d, want %d", w.Code, tt.expectedStatus)
			}

			if tt.serviceErr != nil {
				var got map[string]map[string]string
				if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
					t.Fatal(err)
				}
				want := map[string]map[string]string{
					"error": {"code": httpx.CodeInternalError, "message": "internal server error"},
				}
				if !reflect.DeepEqual(got, want) {
					t.Errorf("response = %v, want %v", got, want)
				}
				return
			}

			var got struct {
				Data []account.Account `json:"data"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got.Data, tt.accounts) {
				t.Errorf("accounts = %+v, want %+v", got.Data, tt.accounts)
			}
			if gotCache := w.Header().Get("Cache-Control"); gotCache == "" {
				t.Error("expected Cache-Control header to be set by NoStoreHeader")
			}
		})
	}
}
