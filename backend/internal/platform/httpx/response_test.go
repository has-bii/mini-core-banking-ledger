package httpx_test

import (
	"backend/internal/platform/httpx"
	"backend/internal/platform/validator"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestResponses(t *testing.T) {
	t.Parallel()
	fields := validator.ErrorField{"name": {"required"}}
	for _, tt := range []struct {
		name, contentType, body string
		status                  int
		send                    func(http.ResponseWriter) error
	}{
		{
			name: "success", status: http.StatusCreated, contentType: "application/json",
			body: `{"data":{"id":1},"message":"created"}`,
			send: func(w http.ResponseWriter) error {
				return httpx.SendSuccess(w, http.StatusCreated, httpx.NewSuccessResponse("created", map[string]int{"id": 1}))
			},
		},
		{
			name: "success without message", status: http.StatusOK, contentType: "application/json",
			body: `{"data":null}`,
			send: func(w http.ResponseWriter) error {
				return httpx.SendSuccess(w, http.StatusOK, httpx.NewSuccessResponse("", nil))
			},
		},
		{
			name: "error without fields", status: http.StatusBadRequest, contentType: "application/problem+json",
			body: `{"type":"about:blank","title":"Bad Request","status":400,"detail":"invalid input","instance":"/accounts"}`,
			send: func(w http.ResponseWriter) error {
				return httpx.SendError(w, httpx.NewErrorResponse("Bad Request", http.StatusBadRequest, "invalid input", "/accounts", &validator.ErrorValidation{}))
			},
		},
		{
			name: "error with fields", status: http.StatusUnprocessableEntity, contentType: "application/problem+json",
			body: `{"type":"about:blank","title":"Invalid","status":422,"detail":"invalid input","instance":"/accounts","errors":{"name":["required"]}}`,
			send: func(w http.ResponseWriter) error {
				return httpx.SendError(w, httpx.NewErrorResponse("Invalid", http.StatusUnprocessableEntity, "invalid input", "/accounts", &validator.ErrorValidation{Errors: &fields}))
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			if err := tt.send(w); err != nil {
				t.Fatal(err)
			}
			if w.Code != tt.status || w.Header().Get("Content-Type") != tt.contentType {
				t.Fatalf("status/content type = %d, %q; want %d, %q", w.Code, w.Header().Get("Content-Type"), tt.status, tt.contentType)
			}
			var got, want any
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(tt.body), &want); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("body = %s, want %s", w.Body.String(), tt.body)
			}
			if !strings.HasSuffix(w.Body.String(), "\n") {
				t.Error("body must end with a newline")
			}
		})
	}
}

func TestErrorResponseWithoutValidationErrors(t *testing.T) {
	t.Parallel()
	w := httptest.NewRecorder()
	response := httpx.NewErrorResponse("Unauthorized", http.StatusUnauthorized, "Session is expired or invalid", "/accounts", nil)
	if err := httpx.SendError(w, response); err != nil {
		t.Fatal(err)
	}
	if w.Code != http.StatusUnauthorized || w.Header().Get("Content-Type") != "application/problem+json" {
		t.Fatalf("response = %d %v", w.Code, w.Header())
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if _, ok := body["errors"]; ok {
		t.Fatalf("unexpected validation errors: %s", w.Body.String())
	}
	if body["detail"] != "Session is expired or invalid" || body["instance"] != "/accounts" {
		t.Fatalf("unexpected problem details: %s", w.Body.String())
	}
}

func TestResponsesRejectBeforeWriting(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		send func(http.ResponseWriter) error
	}{
		{"nil success", func(w http.ResponseWriter) error { return httpx.SendSuccess(w, http.StatusOK, nil) }},
		{"nil error", func(w http.ResponseWriter) error { return httpx.SendError(w, nil) }},
		{"marshal failure", func(w http.ResponseWriter) error {
			return httpx.SendSuccess(w, http.StatusOK, httpx.NewSuccessResponse("", make(chan int)))
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w := &responseWriter{ResponseRecorder: httptest.NewRecorder()}
			if err := tt.send(w); err == nil {
				t.Fatal("expected an error")
			}
			if w.wrote || len(w.Header()) != 0 || w.Body.Len() != 0 {
				t.Fatal("response was modified before returning an error")
			}
		})
	}
}

func TestResponsesReturnWriterError(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		send func(http.ResponseWriter) error
	}{
		{"success", func(w http.ResponseWriter) error {
			return httpx.SendSuccess(w, http.StatusOK, httpx.NewSuccessResponse("", nil))
		}},
		{"error", func(w http.ResponseWriter) error {
			return httpx.SendError(w, httpx.NewErrorResponse("Bad Request", http.StatusBadRequest, "", "", &validator.ErrorValidation{}))
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			want := errors.New("write failed")
			w := &responseWriter{ResponseRecorder: httptest.NewRecorder(), err: want}
			if err := tt.send(w); !errors.Is(err, want) {
				t.Fatalf("error = %v, want %v", err, want)
			}
		})
	}
}

type responseWriter struct {
	*httptest.ResponseRecorder
	wrote bool
	err   error
}

func (w *responseWriter) WriteHeader(status int) {
	w.wrote = true
	w.ResponseRecorder.WriteHeader(status)
}

func (w *responseWriter) Write(body []byte) (int, error) {
	w.wrote = true
	if w.err != nil {
		return 0, w.err
	}
	return w.ResponseRecorder.Write(body)
}
