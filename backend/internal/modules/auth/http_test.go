package auth

import (
	"backend/internal/modules/user"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/jackc/pgx/v5"
)

func TestRegisterHTTP(t *testing.T) {
	const validBody = `{"name":"Alice","email":"alice@example.com","password":"password123"}`
	for _, tc := range []struct {
		name       string
		body       string
		lookupErr  error
		wantStatus int
		wantError  string
	}{
		{name: "invalid JSON", body: `{`, wantStatus: http.StatusBadRequest, wantError: "Invalid JSON body"},
		{name: "missing fields", body: `{}`, wantStatus: http.StatusBadRequest, wantError: "name is required\nemail must be a valid email address\npassword must contain at least 8 characters"},
		{name: "short password", body: `{"name":"Alice","email":"alice@example.com","password":"short"}`, wantStatus: http.StatusBadRequest, wantError: "password must contain at least 8 characters"},
		{name: "long password", body: `{"name":"Alice","email":"alice@example.com","password":"` + strings.Repeat("a", 73) + `"}`, wantStatus: http.StatusBadRequest, wantError: "password must not exceed 72 bytes"},
		{name: "normalized input", body: `{"name":" Alice ","email":" alice@example.com ","password":"password123"}`, lookupErr: pgx.ErrNoRows, wantStatus: http.StatusCreated},
		{name: "duplicate email", body: validBody, wantStatus: http.StatusConflict, wantError: "email already taken"},
		{name: "invalid email", body: `{"name":"Alice","email":"invalid","password":"password123"}`, wantStatus: http.StatusBadRequest, wantError: "email must be a valid email address"},
		{name: "internal error", body: validBody, lookupErr: errors.New("private database details"), wantStatus: http.StatusInternalServerError, wantError: "internal error"},
		{name: "new user", body: validBody, lookupErr: pgx.ErrNoRows, wantStatus: http.StatusCreated},
	} {
		t.Run(tc.name, func(t *testing.T) {
			db := &registrationDB{lookupErr: tc.lookupErr}
			h := NewHandler(NewService(user.NewStore(db)), slog.New(slog.NewTextHandler(io.Discard, nil)))
			mux := http.NewServeMux()
			h.Register(mux)
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, httptest.NewRequest(http.MethodPost, "/auth/register", strings.NewReader(tc.body)))
			if response.Code != tc.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", response.Code, tc.wantStatus, response.Body.String())
			}
			if tc.wantError != "" {
				var body struct {
					Error string `json:"error"`
				}
				if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil || body.Error != tc.wantError {
					t.Fatalf("unexpected error response: %s", response.Body.String())
				}
			}
			if tc.wantStatus == http.StatusCreated {
				var body map[string]any
				if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil || body["name"] != "Alice" || body["email"] != "alice@example.com" {
					t.Fatalf("unexpected user response: %s", response.Body.String())
				}
				if _, ok := body["password"]; ok {
					t.Fatal("response exposed credentials")
				}
			}
			if tc.wantStatus == http.StatusBadRequest && db.queries != 0 {
				t.Fatal("invalid input reached the database")
			}
		})
	}
}
