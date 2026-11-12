package httpx_test

import (
	"backend/internal/platform/transport/httpx"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestDecodeJSON_BodyErrorResponse(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name, body, code, message string
		status                    int
	}{
		{"empty", "", "INVALID_JSON", "Request body is required.", 400},
		{"whitespace", " \t\r\n", "INVALID_JSON", "Request body is required.", 400},
		{"malformed", `{`, "INVALID_JSON", "Request body contains malformed JSON.", 400},
		{"null", `null`, "INVALID_JSON", "Request body must be a JSON object.", 400},
		{"array", `[]`, "INVALID_JSON", "Request body must be a JSON object.", 400},
		{"string", `"secret input"`, "INVALID_JSON", "Request body must be a JSON object.", 400},
		{"number", `123`, "INVALID_JSON", "Request body must be a JSON object.", 400},
		{"multiple values", `{} {}`, "INVALID_JSON", "Request body must contain one JSON object.", 400},
		{"trailing garbage", `{} x`, "INVALID_JSON", "Request body contains malformed JSON.", 400},
		{"oversized", `{"email":"` + strings.Repeat("a", 1<<20) + `"}`, "PAYLOAD_TOO_LARGE", "Request body must not exceed 1 MiB.", 413},
		{"oversized trailing whitespace", `{}` + strings.Repeat(" ", 1<<20), "PAYLOAD_TOO_LARGE", "Request body must not exceed 1 MiB.", 413},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			var payload struct {
				Email string `json:"email"`
			}
			err := httpx.DecodeJSON(w, r, &payload)
			if err == nil {
				t.Fatal("DecodeJSON() accepted an invalid body")
			}
			httpx.WriteDecodeError(w, err)
			var got struct {
				Error struct {
					Code, Message string
					Details       json.RawMessage
				}
			}
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if w.Code != tt.status || got.Error.Code != tt.code || got.Error.Message != tt.message || len(got.Error.Details) != 0 {
				t.Errorf("response = %d %s, want %d %s %q without details", w.Code, w.Body.String(), tt.status, tt.code, tt.message)
			}
			if w.Header().Get("Content-Type") != "application/json" {
				t.Error("error response must have JSON content type")
			}
		})
	}
}

func TestDecodeJSON_RequiresObjectForAnyDestination(t *testing.T) {
	t.Parallel()
	for _, body := range []string{`null`, `[]`, `"text"`, `123`, `true`} {
		t.Run(body, func(t *testing.T) {
			t.Parallel()
			var payload any
			w := httptest.NewRecorder()
			err := httpx.DecodeJSON(w, httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body)), &payload)
			if err == nil {
				t.Fatal("DecodeJSON() accepted a non-object body")
			}
			httpx.WriteDecodeError(w, err)
			if w.Code != 400 || !strings.Contains(w.Body.String(), `"message":"Request body must be a JSON object."`) {
				t.Errorf("response = %d %s, want an object-required error", w.Code, w.Body.String())
			}
		})
	}
}

func TestDecodeJSON_ValidObject(t *testing.T) {
	t.Parallel()
	body := `{"email":"alice@example.com"}`
	for _, tt := range []struct{ name, body string }{
		{"normal", body},
		{"surrounding whitespace", " \n" + body + "\r\t "},
		{"exactly 1 MiB", body + strings.Repeat(" ", (1<<20)-len(body))},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var payload struct {
				Email string `json:"email"`
			}
			if err := httpx.DecodeJSON(httptest.NewRecorder(), httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body)), &payload); err != nil {
				t.Fatal(err)
			}
			if payload.Email != "alice@example.com" {
				t.Errorf("email = %q, want alice@example.com", payload.Email)
			}
		})
	}
}

func TestDecodeJSON_FieldErrorResponse(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct{ name, body, want string }{
		{"wrong type", `{"email":123}`, `{"error":{"code":"INVALID_JSON","message":"Field \"email\" must be a string.","details":[{"field":"email","code":"INVALID_TYPE","message":"Field \"email\" must be a string."}]}}`},
		{"no submitted values", `{"email":["secret password"]}`, `{"error":{"code":"INVALID_JSON","message":"Field \"email\" must be a string.","details":[{"field":"email","code":"INVALID_TYPE","message":"Field \"email\" must be a string."}]}}`},
		{"unknown field", `{"balance":100}`, `{"error":{"code":"INVALID_JSON","message":"Field \"balance\" is not allowed.","details":[{"field":"balance","code":"UNKNOWN_FIELD","message":"Field \"balance\" is not allowed."}]}}`},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			var payload struct {
				Email string `json:"email"`
			}
			if err := httpx.DecodeJSON(w, r, &payload); err != nil {
				httpx.WriteDecodeError(w, err)
			} else {
				t.Fatal("DecodeJSON() accepted an invalid field")
			}
			if w.Code != http.StatusBadRequest {
				t.Fatalf("status = %d, want 400", w.Code)
			}
			var got, want any
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(tt.want), &want); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("response = %s, want %s", w.Body.String(), tt.want)
			}
		})
	}
}
