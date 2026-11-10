package account_test

import (
	"backend/internal/modules/account"
	"backend/internal/platform/transport/httpx"
	"backend/internal/platform/transport/httpx/middleware"
	"backend/internal/platform/validation"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestHandler_GetAccounts(t *testing.T) {
	t.Parallel()
	const userID = "01a1c6de-805e-7b0c-a1b1-b851a9c38ccd"

	for _, tt := range []struct {
		name     string
		accounts []account.Account
		err      error
	}{
		{
			name: "returns accounts",
			accounts: []account.Account{
				{ID: "01a1c6de-805e-7b0c-a1b1-b851a9c38cce", UserID: userID, Name: "Main Account", Type: account.Main, Balance: 12345},
			},
		},
		{name: "no accounts", accounts: []account.Account{}},
		{name: "invalid user ID error", err: validation.ErrInvalidUUID},
		{name: "not found error", err: account.ErrNotFound},
		{name: "unexpected error", err: errors.New("secret database credentials")},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			store := &accountStoreStub{accounts: tt.accounts, err: tt.err}
			mux := http.NewServeMux()
			account.NewHandler(account.NewService(store), slog.New(slog.NewTextHandler(io.Discard, nil))).Register(mux)

			var requestCtx context.Context
			h := middleware.GetUser(func(context.Context, string) (string, error) {
				return userID, nil
			})(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				requestCtx = r.Context()
				mux.ServeHTTP(w, r)
			}))
			r := httptest.NewRequest(http.MethodGet, "/accounts", nil).WithContext(t.Context())
			r.AddCookie(&http.Cookie{Name: "session", Value: "test-session"})
			w := httptest.NewRecorder()
			h.ServeHTTP(w, r)

			if store.userID != userID {
				t.Errorf("service received userID = %q, want %q", store.userID, userID)
			}
			if store.ctx != requestCtx {
				t.Error("service did not receive the authenticated request context")
			}

			if tt.err != nil {
				if w.Code != http.StatusInternalServerError {
					t.Fatalf("status = %d, want %d", w.Code, http.StatusInternalServerError)
				}
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

			if w.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
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
		})
	}
}
