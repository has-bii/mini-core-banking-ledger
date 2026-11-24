package httpx_test

import (
	"backend/internal/platform/httpx"
	"backend/internal/platform/logger"
	"backend/internal/platform/validator"
	"encoding/json/v2"
	"errors"
	"fmt"
	"log/slog"
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
		path                    string
		send                    func(http.ResponseWriter, *http.Request)
	}{
		{
			name: "success", status: http.StatusCreated, contentType: "application/json",
			body: `{"data":{"id":1},"message":"created"}`,
			send: func(w http.ResponseWriter, r *http.Request) {
				httpx.SendSuccess(w, r, http.StatusCreated, httpx.SuccessResponse{Data: map[string]int{"id": 1}, Message: "created"})
			},
		},
		{
			name: "success without message", status: http.StatusOK, contentType: "application/json",
			body: `{"data":null}`,
			send: func(w http.ResponseWriter, r *http.Request) {
				httpx.SendSuccess(w, r, http.StatusOK, httpx.SuccessResponse{Data: []int(nil)})
			},
		},
		{
			name: "error without fields", status: http.StatusBadRequest, contentType: "application/problem+json",
			body: `{"type":"about:blank","title":"Bad Request","status":400,"detail":"invalid input","instance":"/wallets"}`,
			send: func(w http.ResponseWriter, r *http.Request) {
				httpx.SendError(w, r, http.StatusBadRequest, "invalid input", &validator.ErrorValidation{})
			},
		},
		{
			name: "error with invalid UTF-8 in instance", status: http.StatusUnauthorized, contentType: "application/problem+json",
			body: `{"type":"about:blank","title":"Unauthorized","status":401,"detail":"invalid session","instance":"/wallets/\ufffd"}`,
			path: "/wallets/%FF",
			send: func(w http.ResponseWriter, r *http.Request) {
				httpx.SendError(w, r, http.StatusUnauthorized, "invalid session", nil)
			},
		},
		{
			name: "error with empty fields map", status: http.StatusBadRequest, contentType: "application/problem+json",
			body: `{"type":"about:blank","title":"Bad Request","status":400,"detail":"invalid input","instance":"/wallets","errors":{}}`,
			send: func(w http.ResponseWriter, r *http.Request) {
				empty := validator.ErrorField{}
				httpx.SendError(w, r, http.StatusBadRequest, "invalid input", &validator.ErrorValidation{Errors: &empty})
			},
		},
		{
			name: "error with fields", status: http.StatusUnprocessableEntity, contentType: "application/problem+json",
			body: `{"type":"about:blank","title":"Unprocessable Entity","status":422,"detail":"invalid input","instance":"/wallets","errors":{"name":["required"]}}`,
			send: func(w http.ResponseWriter, r *http.Request) {
				httpx.SendError(w, r, http.StatusUnprocessableEntity, "invalid input", &validator.ErrorValidation{Errors: &fields})
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			path := tt.path
			if path == "" {
				path = "/wallets"
			}
			r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, path, nil)
			tt.send(w, r)
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
	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/wallets", nil)
	httpx.SendError(w, r, http.StatusUnauthorized, "Session is expired or invalid", nil)
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
	if body["detail"] != "Session is expired or invalid" || body["instance"] != "/wallets" {
		t.Fatalf("unexpected problem details: %s", w.Body.String())
	}
}

func TestResponsesFallBackOnEncodingError(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name string
		send func(http.ResponseWriter, *http.Request)
	}{
		{"success", func(w http.ResponseWriter, r *http.Request) {
			httpx.SendSuccess(w, r, http.StatusOK, httpx.SuccessResponse{Data: make(chan int)})
		}},
		{"error", func(w http.ResponseWriter, r *http.Request) {
			httpx.SendError(w, r, http.StatusBadRequest, "invalid \xff detail", nil)
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var logs strings.Builder
			ctx := logger.WithContext(t.Context(), slog.New(slog.NewTextHandler(&logs, nil)))
			r := httptest.NewRequestWithContext(ctx, http.MethodPost, "/resources", nil)
			w := httptest.NewRecorder()
			w.Header().Set("Set-Cookie", "session=staged")
			httpx.NoStoreHeader(w)

			tt.send(w, r)

			if w.Code != http.StatusInternalServerError || w.Header().Get("Content-Type") != "application/problem+json" {
				t.Fatalf("response = %d %v, want status 500 and problem JSON", w.Code, w.Header())
			}
			if w.Header().Get("Set-Cookie") != "" || w.Header().Get("Cache-Control") != "no-store" {
				t.Errorf("unexpected fallback headers: %v", w.Header())
			}
			var got httpx.ErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if got.Type != "about:blank" || got.Title != "Internal Server Error" || got.Status != 500 ||
				got.Detail != "internal error" || got.Instance != "" || got.Errors != nil {
				t.Errorf("unexpected fallback problem details: %s", w.Body.String())
			}
			if !strings.Contains(logs.String(), "encode JSON response") {
				t.Errorf("encoding failure missing from request log: %s", logs.String())
			}
		})
	}
}

func TestResponsesLogWriterErrorWithoutRetry(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name   string
		status int
		send   func(http.ResponseWriter, *http.Request)
	}{
		{"success", http.StatusOK, func(w http.ResponseWriter, r *http.Request) {
			httpx.SendSuccess(w, r, http.StatusOK, httpx.SuccessResponse{})
		}},
		{"error", http.StatusBadRequest, func(w http.ResponseWriter, r *http.Request) {
			httpx.SendError(w, r, http.StatusBadRequest, "", nil)
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var logs strings.Builder
			ctx := logger.WithContext(t.Context(), slog.New(slog.NewTextHandler(&logs, nil)))
			r := httptest.NewRequestWithContext(ctx, http.MethodGet, "/resources", nil)
			cause := errors.New("writer failed")
			w := &failingResponseWriter{ResponseRecorder: httptest.NewRecorder(), err: cause}

			tt.send(w, r)

			if w.Code != tt.status || w.writes != 1 {
				t.Errorf("status = %d, writes = %d; want %d and one write", w.Code, w.writes, tt.status)
			}
			if !strings.Contains(logs.String(), "write JSON response") || !strings.Contains(logs.String(), cause.Error()) {
				t.Errorf("writer failure missing from request log: %s", logs.String())
			}
		})
	}
}

func TestSendRequestError(t *testing.T) {
	t.Parallel()
	fields := validator.ErrorField{"name": {"required"}}
	validation := &validator.ErrorValidation{Errors: &fields}
	conflict := errors.New("private conflict cause")
	unexpected := errors.New("backend unavailable")
	for _, tt := range []struct {
		name, detail string
		err          error
		status       int
		fields       *validator.ErrorField
	}{
		{"decode takes precedence", "body too large", errors.Join(httpx.NewHTTPError(413, "body too large"), validation, conflict), 413, nil},
		{"validation takes precedence", "One or more fields are invalid. Check the field errors for details.", errors.Join(validation, conflict), 400, &fields},
		{"wrapped domain error", "resource already exists", fmt.Errorf("save: %w", conflict), 409, nil},
		{"unexpected error", "internal error", unexpected, 500, nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var logs strings.Builder
			log := slog.New(slog.NewTextHandler(&logs, nil))
			ctx := logger.WithContext(t.Context(), log)
			r := httptest.NewRequestWithContext(ctx, http.MethodPost, "/resources?debug=true", nil)
			w := httptest.NewRecorder()
			httpx.SendRequestError(w, r, "test", tt.err, func(err error) (int, string) {
				if errors.Is(err, conflict) {
					return http.StatusConflict, "resource already exists"
				}
				return 0, ""
			})
			if w.Code != tt.status || w.Header().Get("Content-Type") != "application/problem+json" {
				t.Fatalf("response = %d %v, want status %d and problem JSON", w.Code, w.Header(), tt.status)
			}
			var got httpx.ErrorResponse
			if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
				t.Fatal(err)
			}
			if got.Status != tt.status || got.Detail != tt.detail || got.Instance != "/resources" || !reflect.DeepEqual(got.Errors, tt.fields) {
				t.Errorf("unexpected problem details: %s", w.Body.String())
			}
			if tt.status == 500 {
				if !strings.Contains(logs.String(), unexpected.Error()) {
					t.Errorf("internal error cause missing from log: %s", logs.String())
				}
			} else if logs.Len() != 0 {
				t.Errorf("client error was logged: %s", logs.String())
			}
		})
	}
}

type failingResponseWriter struct {
	*httptest.ResponseRecorder
	writes int
	err    error
}

func (w *failingResponseWriter) Write([]byte) (int, error) {
	w.writes++
	return 0, w.err
}
