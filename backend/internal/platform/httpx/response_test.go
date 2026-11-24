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
			name: "validation error without fields", status: http.StatusBadRequest, contentType: "application/problem+json",
			body: `{"type":"about:blank","title":"Bad Request","status":400,"detail":"One or more fields are invalid. Check the field errors for details.","instance":"/wallets"}`,
			send: func(w http.ResponseWriter, r *http.Request) {
				httpx.SendError(w, r, &validator.ErrorValidation{})
			},
		},
		{
			name: "error with invalid UTF-8 in instance", status: http.StatusUnauthorized, contentType: "application/problem+json",
			body: `{"type":"about:blank","title":"Unauthorized","status":401,"detail":"invalid session","instance":"/wallets/\ufffd"}`,
			path: "/wallets/%FF",
			send: func(w http.ResponseWriter, r *http.Request) {
				httpx.SendError(w, r, httpx.NewHTTPError(http.StatusUnauthorized, "invalid session"))
			},
		},
		{
			name: "error with empty fields map", status: http.StatusBadRequest, contentType: "application/problem+json",
			body: `{"type":"about:blank","title":"Bad Request","status":400,"detail":"One or more fields are invalid. Check the field errors for details.","instance":"/wallets","errors":{}}`,
			send: func(w http.ResponseWriter, r *http.Request) {
				empty := validator.ErrorField{}
				httpx.SendError(w, r, &validator.ErrorValidation{Errors: &empty})
			},
		},
		{
			name: "error with fields", status: http.StatusBadRequest, contentType: "application/problem+json",
			body: `{"type":"about:blank","title":"Bad Request","status":400,"detail":"One or more fields are invalid. Check the field errors for details.","instance":"/wallets","errors":{"name":["required"]}}`,
			send: func(w http.ResponseWriter, r *http.Request) {
				httpx.SendError(w, r, &validator.ErrorValidation{Errors: &fields})
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
	httpx.SendError(w, r, httpx.NewHTTPError(http.StatusUnauthorized, "Session is expired or invalid"))
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
		name, target, instance string
		send                   func(http.ResponseWriter, *http.Request)
	}{
		{"success", "/resources?limit=10", "/resources", func(w http.ResponseWriter, r *http.Request) {
			httpx.SendSuccess(w, r, http.StatusOK, httpx.SuccessResponse{Data: make(chan int)})
		}},
		{"error", "/resources/%22%0B%FF%FE", "/resources/\"\v\uFFFD", func(w http.ResponseWriter, r *http.Request) {
			httpx.SendError(w, r, httpx.NewHTTPError(http.StatusBadRequest, "invalid \xff detail"))
		}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var logs strings.Builder
			ctx := logger.WithContext(t.Context(), slog.New(slog.NewTextHandler(&logs, nil)))
			r := httptest.NewRequestWithContext(ctx, http.MethodPost, tt.target, nil)
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
				got.Detail != "internal error" || got.Instance != tt.instance || got.Errors != nil {
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
			httpx.SendError(w, r, httpx.NewHTTPError(http.StatusBadRequest, ""))
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

func TestSendError(t *testing.T) {
	t.Parallel()
	fields := validator.ErrorField{"name": {"required"}}
	validation := &validator.ErrorValidation{Errors: &fields}
	unavailable := errors.New("backend unavailable")
	for _, tt := range []struct {
		name, detail string
		err          error
		status       int
		fields       *validator.ErrorField
		wantLog      string
	}{
		{"transport takes precedence", "body too large", errors.Join(validation, httpx.NewHTTPError(413, "body too large")), 413, nil, ""},
		{"wrapped validation error", "One or more fields are invalid. Check the field errors for details.", fmt.Errorf("validate: %w", validation), 400, &fields, ""},
		{"client error with cause", "resource already exists", httpx.WrapHTTPError(409, "resource already exists", errors.New("private conflict cause")), 409, nil, ""},
		{"unexpected error", "internal error", unavailable, 500, nil, "backend unavailable"},
		{"explicit server error with cause", "please retry later", httpx.WrapHTTPError(503, "please retry later", unavailable), 503, nil, "backend unavailable"},
		{"explicit server error without cause", "Unexpected error has occurred", httpx.NewHTTPError(500, "Unexpected error has occurred"), 500, nil, ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var logs strings.Builder
			log := slog.New(slog.NewTextHandler(&logs, nil))
			ctx := logger.WithContext(t.Context(), log)
			r := httptest.NewRequestWithContext(ctx, http.MethodPost, "/resources?debug=true", nil)
			w := httptest.NewRecorder()
			httpx.SendError(w, r, tt.err)
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
			if tt.wantLog == "" {
				if logs.Len() != 0 {
					t.Errorf("unexpected request failure log: %s", logs.String())
				}
				return
			}
			if strings.Count(logs.String(), "\n") != 1 || !strings.Contains(logs.String(), `msg="request failed"`) ||
				!strings.Contains(logs.String(), tt.wantLog) || !strings.Contains(logs.String(), "method=POST path=/resources") {
				t.Errorf("want one request failure log with cause, method, and path; got: %s", logs.String())
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
