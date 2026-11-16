package middleware_test

import (
	"backend/internal/platform/v2/httpx"
	"backend/internal/platform/v2/middleware"
	"backend/internal/platform/v2/validator"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleErrors(t *testing.T) {
	t.Parallel()
	fields := []validator.ErrorField{{Field: "name", Message: []string{"required"}}}
	for _, tt := range []struct {
		name   string
		err    error
		status int
		detail string
		logged bool
		fields bool
	}{
		{name: "bad body", err: httpx.BadRequest(400, "invalid JSON"), status: 400, detail: "invalid JSON"},
		{name: "wrapped oversized body", err: fmt.Errorf("decode: %w", httpx.BadRequest(413, "body too large")), status: 413, detail: "body too large"},
		{name: "wrapped validation", err: fmt.Errorf("validate: %w", validator.NewErrorValidation(&fields)), status: 422, detail: "One or more fields are invalid. Check the field errors for details.", fields: true},
		{name: "unexpected", err: errors.New("private database failure"), status: 500, detail: "Unexpected error has occurred", logged: true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var logs bytes.Buffer
			log := slog.New(slog.NewJSONHandler(&logs, nil))
			rr := httptest.NewRecorder()
			middleware.HandleErrors(func(http.ResponseWriter, *http.Request) error {
				return tt.err
			}, log).ServeHTTP(rr, httptest.NewRequest(http.MethodPost, "/accounts?secret=hidden", nil))
			if rr.Code != tt.status || rr.Header().Get("Content-Type") != "application/problem+json" {
				t.Fatalf("response: status=%d headers=%v", rr.Code, rr.Header())
			}
			var response httpx.ErrorResponse
			if err := json.Unmarshal(rr.Body.Bytes(), &response); err != nil {
				t.Fatal(err)
			}
			if response.Detail != tt.detail || response.Status != tt.status || response.Instance != "/accounts" || response.Title != http.StatusText(tt.status) {
				t.Fatalf("unexpected response: %+v", response)
			}
			if tt.fields && (response.Error == nil || len(*response.Error) != 1 || (*response.Error)[0].Field != "name") {
				t.Fatalf("missing field errors: %+v", response)
			}
			if (logs.Len() > 0) != tt.logged {
				t.Fatalf("unexpected logs: %s", &logs)
			}
			if tt.logged && (!strings.Contains(logs.String(), tt.err.Error()) || !strings.Contains(logs.String(), `"method":"POST"`) || !strings.Contains(logs.String(), `"path":"/accounts"`)) {
				t.Fatalf("missing error context: %s", &logs)
			}
		})
	}
}

func TestHandleErrorsCommittedResponse(t *testing.T) {
	t.Parallel()
	for _, explicit := range []bool{false, true} {
		t.Run(fmt.Sprintf("explicit_header_%t", explicit), func(t *testing.T) {
			t.Parallel()
			var logs bytes.Buffer
			rr := httptest.NewRecorder()
			middleware.HandleErrors(func(w http.ResponseWriter, r *http.Request) error {
				if explicit {
					w.WriteHeader(http.StatusAccepted)
				} else {
					_, _ = w.Write([]byte("original"))
				}
				return errors.New("late failure")
			}, slog.New(slog.NewJSONHandler(&logs, nil))).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
			if explicit {
				if rr.Code != http.StatusAccepted || rr.Body.Len() != 0 {
					t.Fatalf("response modified: %d %q", rr.Code, rr.Body.String())
				}
			} else if rr.Code != http.StatusOK || rr.Body.String() != "original" {
				t.Fatalf("response modified: %d %q", rr.Code, rr.Body.String())
			}
			if !strings.Contains(logs.String(), "late failure") {
				t.Fatalf("missing failure log: %s", &logs)
			}
		})
	}
}

func TestHandleErrorsSuccess(t *testing.T) {
	t.Parallel()
	var logs bytes.Buffer
	rr := httptest.NewRecorder()
	middleware.HandleErrors(func(w http.ResponseWriter, r *http.Request) error {
		w.WriteHeader(http.StatusNoContent)
		return nil
	}, slog.New(slog.NewJSONHandler(&logs, nil))).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Code != http.StatusNoContent || rr.Body.Len() != 0 || logs.Len() != 0 {
		t.Fatalf("unexpected response or logs: %d %q %s", rr.Code, rr.Body.String(), &logs)
	}
}

func TestHandleErrorsReusesRecovererWriter(t *testing.T) {
	t.Parallel()
	var logs bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&logs, nil))
	rr := httptest.NewRecorder()
	handler := middleware.Recoverer(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		middleware.HandleErrors(func(got http.ResponseWriter, r *http.Request) error {
			if got != w {
				t.Error("HandleErrors wrapped the existing response writer")
			}
			got.WriteHeader(http.StatusAccepted)
			return errors.New("late failure")
		}, log).ServeHTTP(w, r)
	}))
	handler.ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
	if rr.Code != http.StatusAccepted || rr.Body.Len() != 0 {
		t.Fatalf("response modified: %d %q", rr.Code, rr.Body.String())
	}
	if !strings.Contains(logs.String(), "late failure") {
		t.Fatalf("missing failure log: %s", &logs)
	}
}

type failingErrorWriter struct {
	header http.Header
	writes int
}

func (w *failingErrorWriter) Header() http.Header { return w.header }
func (w *failingErrorWriter) WriteHeader(int)     {}
func (w *failingErrorWriter) Write([]byte) (int, error) {
	w.writes++
	return 0, errors.New("connection failed")
}

func TestHandleErrorsWriteFailure(t *testing.T) {
	t.Parallel()
	var logs bytes.Buffer
	w := &failingErrorWriter{header: make(http.Header)}
	middleware.HandleErrors(func(http.ResponseWriter, *http.Request) error {
		return httpx.BadRequest(http.StatusBadRequest, "invalid JSON")
	}, slog.New(slog.NewJSONHandler(&logs, nil))).ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))
	if w.writes != 1 || !strings.Contains(logs.String(), "connection failed") {
		t.Fatalf("writes=%d logs=%s", w.writes, &logs)
	}
}
