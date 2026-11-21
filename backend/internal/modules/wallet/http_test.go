package wallet_test

import (
	"backend/internal/modules/wallet"
	"backend/internal/platform/httpx"
	"backend/internal/platform/middleware"
	"backend/internal/platform/validator"
	"context"
	"encoding/json/v2"
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

const walletTestUserID = "01a1c6de-805e-7b0c-a1b1-b851a9c38ccd"

type walletServiceStub struct {
	ctx        context.Context
	id, userID string
	data       *wallet.CreateWalletDTO
	wallet     wallet.Wallet
	wallets    []wallet.Wallet
	err        error
}

func (s *walletServiceStub) CreateUserWallet(ctx context.Context, userID string, data *wallet.CreateWalletDTO) (wallet.Wallet, error) {
	s.ctx, s.userID, s.data = ctx, userID, data
	return s.wallet, s.err
}

func (s *walletServiceStub) GetUserWallet(ctx context.Context, id, userID string) (wallet.Wallet, error) {
	s.ctx, s.id, s.userID = ctx, id, userID
	return s.wallet, s.err
}

func (s *walletServiceStub) GetUserWallets(ctx context.Context, userID string) ([]wallet.Wallet, error) {
	s.ctx, s.userID = ctx, userID
	return s.wallets, s.err
}

func walletTestHandler(t *testing.T, svc *walletServiceStub) http.Handler {
	t.Helper()
	mux := http.NewServeMux()
	wallet.NewHandler(svc, slog.New(slog.NewTextHandler(io.Discard, nil))).Register(mux)
	return middleware.GetUserFromSession(func(context.Context, string) (string, error) {
		return walletTestUserID, nil
	})(mux)
}

func assertWalletProblem(t *testing.T, w *httptest.ResponseRecorder, status int, detail, instance string, fields validator.ErrorField) {
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

func TestHandler_PreservesVerifiedIdentityAndContext(t *testing.T) {
	t.Parallel()
	svc := &walletServiceStub{wallets: []wallet.Wallet{}}
	type requestKey struct{}
	ctx := context.WithValue(t.Context(), requestKey{}, "request-value")
	ctx = middleware.WithUserID(ctx, "native-user")
	r := httptest.NewRequest(http.MethodGet, "/wallets", nil).WithContext(ctx)
	w := httptest.NewRecorder()
	walletTestHandler(t, svc).ServeHTTP(w, r)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", w.Code, http.StatusOK, w.Body.String())
	}
	if svc.userID != "native-user" || svc.ctx == nil || svc.ctx.Value(requestKey{}) != "request-value" {
		t.Fatalf("identity or request context lost: %+v", svc)
	}
}

func TestHandler_CreateWallet(t *testing.T) {
	t.Parallel()
	want := wallet.Wallet{ID: "01a1c6de-805e-7b0c-a1b1-b851a9c38cce", UserID: walletTestUserID, Name: "Travel", Type: wallet.Savings}
	svc := &walletServiceStub{wallet: want}
	r := httptest.NewRequest(http.MethodPost, "/wallets", strings.NewReader(`{"name":" Travel ","wallet_type":" savings "}`)).WithContext(t.Context())
	r.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
	w := httptest.NewRecorder()
	walletTestHandler(t, svc).ServeHTTP(w, r)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d: %s", w.Code, http.StatusCreated, w.Body.String())
	}
	wantInput := wallet.CreateWalletDTO{Name: "Travel", WalletType: "savings"}
	if svc.data == nil || *svc.data != wantInput || svc.userID != walletTestUserID {
		t.Errorf("service received input %+v for user %q, want %+v for %q", svc.data, svc.userID, wantInput, walletTestUserID)
	}
	var got struct {
		Data wallet.Wallet `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatal(err)
	}
	if got.Data != want {
		t.Errorf("response data = %+v, want %+v", got.Data, want)
	}
}

func TestHandler_CreateWalletRejectsInvalidPayload(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, body, detail string
		status             int
		fields             validator.ErrorField
	}{
		{name: "malformed JSON", body: `{`, status: 400, detail: "body contains badly-formed JSON"},
		{name: "wrong field type", body: `{"name":123,"wallet_type":"savings"}`, status: 400, detail: `body contains an invalid value for field "name" (expected string)`},
		{name: "client supplied owner", body: `{"name":"Travel","wallet_type":"savings","user_id":"another-user"}`, status: 400, detail: `body contains unknown field "user_id"`},
		{name: "client supplied balance", body: `{"name":"Travel","wallet_type":"savings","balance":100}`, status: 400, detail: `body contains unknown field "balance"`},
		{name: "oversized body", body: `{"name":"` + strings.Repeat("A", 1<<20) + `","wallet_type":"savings"}`, status: 413, detail: "body must not be larger than 1048576 bytes"},
		{name: "missing fields", body: `{}`, status: 400, detail: "One or more fields are invalid. Check the field errors for details.", fields: validator.ErrorField{
			"name": {"This field is required"}, "wallet_type": {"This field is required"},
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			r := httptest.NewRequest(http.MethodPost, "/wallets", strings.NewReader(tt.body)).WithContext(t.Context())
			r.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
			w := httptest.NewRecorder()
			svc := &walletServiceStub{}
			walletTestHandler(t, svc).ServeHTTP(w, r)
			assertWalletProblem(t, w, tt.status, tt.detail, "/wallets", tt.fields)
			if svc.ctx != nil {
				t.Error("invalid payload reached service")
			}
		})
	}
}

func TestHandler_CreateWalletErrors(t *testing.T) {
	t.Parallel()
	fields := validator.ErrorField{"name": {"This field is required"}}
	for _, tt := range []struct {
		name            string
		unauthenticated bool
		serviceErr      error
		status          int
		detail          string
		fields          validator.ErrorField
	}{
		{name: "unauthenticated", unauthenticated: true, status: 401, detail: "Session is expired or invalid"},
		{name: "missing owner", serviceErr: fmt.Errorf("create: %w", wallet.ErrNotFound), status: 404, detail: "wallet not found"},
		{name: "invalid user ID", serviceErr: fmt.Errorf("create: %w", wallet.ErrInvalidUserID), status: 400, detail: "invalid user ID"},
		{name: "wrapped validation error", serviceErr: fmt.Errorf("create: %w", &validator.ErrorValidation{Errors: &fields}), status: 400, detail: "One or more fields are invalid. Check the field errors for details.", fields: fields},
		{name: "unexpected error", serviceErr: errors.New("database password must not leak"), status: 500, detail: "internal error"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &walletServiceStub{err: tt.serviceErr}
			r := httptest.NewRequest(http.MethodPost, "/wallets", strings.NewReader(`{"name":"Travel","wallet_type":"savings"}`)).WithContext(t.Context())
			if !tt.unauthenticated {
				r.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
			}
			w := httptest.NewRecorder()
			walletTestHandler(t, svc).ServeHTTP(w, r)
			assertWalletProblem(t, w, tt.status, tt.detail, "/wallets", tt.fields)
			if tt.unauthenticated && svc.ctx != nil {
				t.Error("unauthenticated request reached service")
			}
		})
	}
}

func TestHandler_GetWallet(t *testing.T) {
	t.Parallel()
	const id = "01a1c6de-805e-7b0c-a1b1-b851a9c38cce"
	want := wallet.Wallet{ID: id, UserID: walletTestUserID, Name: "Main Wallet", Type: wallet.Main, Balance: 12345}
	for _, tt := range []struct {
		name, id, detail string
		unauthenticated  bool
		serviceErr       error
		status           int
	}{
		{name: "returns wallet", id: id, status: 200},
		{name: "invalid wallet ID", id: "invalid", serviceErr: fmt.Errorf("lookup: %w", wallet.ErrInvalidWalletID), status: 400, detail: "invalid wallet ID"},
		{name: "invalid user ID", id: id, serviceErr: fmt.Errorf("lookup: %w", wallet.ErrInvalidUserID), status: 400, detail: "invalid user ID"},
		{name: "missing or unowned wallet", id: id, serviceErr: fmt.Errorf("lookup: %w", wallet.ErrNotFound), status: 404, detail: "wallet not found"},
		{name: "unexpected error", id: id, serviceErr: errors.New("database connection timeout"), status: 500, detail: "internal error"},
		{name: "unauthenticated", id: id, unauthenticated: true, status: 401, detail: "Session is expired or invalid"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &walletServiceStub{wallet: want, err: tt.serviceErr}
			r := httptest.NewRequest(http.MethodGet, "/wallets/"+tt.id, nil).WithContext(t.Context())
			if !tt.unauthenticated {
				r.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
			}
			w := httptest.NewRecorder()
			walletTestHandler(t, svc).ServeHTTP(w, r)
			if tt.unauthenticated {
				if svc.ctx != nil {
					t.Error("unauthenticated request reached service")
				}
			} else if svc.id != tt.id || svc.userID != walletTestUserID {
				t.Error("service did not receive requested wallet ID and authenticated user ID")
			}
			if tt.status != http.StatusOK {
				assertWalletProblem(t, w, tt.status, tt.detail, r.URL.Path, nil)
				return
			}
			var got struct {
				Data wallet.Wallet `json:"data"`
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

func TestHandler_GetWallets(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name       string
		wallets    []wallet.Wallet
		serviceErr error
		status     int
		detail     string
	}{
		{name: "returns wallets", wallets: []wallet.Wallet{{ID: "01a1c6de-805e-7b0c-a1b1-b851a9c38cce", UserID: walletTestUserID, Name: "Main Wallet", Type: wallet.Main, Balance: 12345}}, status: 200},
		{name: "empty list", wallets: []wallet.Wallet{}, status: 200},
		{name: "invalid user ID", serviceErr: wallet.ErrInvalidUserID, status: 400, detail: "invalid user ID"},
		{name: "unexpected error", serviceErr: errors.New("database connection timeout"), status: 500, detail: "internal error"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := &walletServiceStub{wallets: tt.wallets, err: tt.serviceErr}
			r := httptest.NewRequest(http.MethodGet, "/wallets", nil).WithContext(t.Context())
			r.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
			w := httptest.NewRecorder()
			walletTestHandler(t, svc).ServeHTTP(w, r)
			if svc.userID != walletTestUserID {
				t.Error("service did not receive authenticated user ID")
			}
			if tt.status != http.StatusOK {
				assertWalletProblem(t, w, tt.status, tt.detail, "/wallets", nil)
				return
			}
			var got struct {
				Data []wallet.Wallet `json:"data"`
			}
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if w.Code != 200 || w.Header().Get("Content-Type") != "application/json" || w.Header().Get("Cache-Control") != "no-store" || !reflect.DeepEqual(got.Data, tt.wallets) {
				t.Fatalf("response = %d %s, want %+v", w.Code, w.Body.String(), tt.wallets)
			}
		})
	}
}
