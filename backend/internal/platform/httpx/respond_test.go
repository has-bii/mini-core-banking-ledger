package httpx_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"backend/internal/platform/httpx"
)

func TestDecodeJSON(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		body    string
		wantErr bool
	}{
		{"valid", `{"name":"Alice"}`, false},
		{"trailing whitespace", "{\"name\":\"Alice\"} \n", false},
		{"empty", "", true},
		{"malformed", `{"name":`, true},
		{"unknown field", `{"unexpected":true}`, true},
		{"wrong type", `{"name":1}`, true},
		{"multiple values", `{"name":"Alice"} {}`, true},
		{"trailing garbage", `{"name":"Alice"} invalid`, true},
		{"too large", `{"name":"` + strings.Repeat("a", 1<<20) + `"}`, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.body))
			var dst struct {
				Name string `json:"name"`
			}
			err := httpx.DecodeJSON(httptest.NewRecorder(), r, &dst)
			if (err != nil) != tt.wantErr {
				t.Fatalf("DecodeJSON() error = %v, wantErr = %v", err, tt.wantErr)
			}
			if !tt.wantErr && dst.Name != "Alice" {
				t.Fatalf("decoded name = %q", dst.Name)
			}
		})
	}
}

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	httpx.WriteJSON(w, http.StatusCreated, map[string]string{"id": "123"})
	if w.Code != http.StatusCreated || w.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("status = %d, headers = %v", w.Code, w.Header())
	}
	if strings.TrimSpace(w.Body.String()) != `{"data":{"id":"123"}}` {
		t.Fatalf("unexpected success body: %s", w.Body)
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	httpx.WriteError(w, http.StatusBadRequest, "VALIDATION_ERROR", "invalid input")
	if w.Code != http.StatusBadRequest || w.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("status = %d, headers = %v", w.Code, w.Header())
	}
	var body map[string]map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if body["error"]["code"] != "VALIDATION_ERROR" || body["error"]["message"] != "invalid input" || len(body) != 1 || len(body["error"]) != 2 {
		t.Fatalf("unexpected error body: %v", body)
	}
}
