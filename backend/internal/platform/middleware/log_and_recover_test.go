package middleware_test

import (
	"backend/internal/platform/middleware"

	"bytes"
	"encoding/json/v2"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestLogAndRecover(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		handler      http.HandlerFunc
		wantStatus   int
		wantBody     string
		wantPanic    any
		wantPanicLog bool
	}{
		{name: "no writes defaults to 200", handler: func(http.ResponseWriter, *http.Request) {}, wantStatus: http.StatusOK},
		{name: "body implies 200", handler: func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("hello"))
		}, wantStatus: http.StatusOK, wantBody: "hello"},
		{name: "explicit status", handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, wantStatus: http.StatusNoContent},
		{name: "normal response passes through", handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			_, _ = w.Write([]byte("created"))
		}, wantStatus: http.StatusCreated, wantBody: "created"},
		{name: "first final status wins", handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			w.WriteHeader(http.StatusInternalServerError)
		}, wantStatus: http.StatusCreated},
		{name: "panic after final headers aborts", handler: func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusAccepted)
			panic("private panic detail")
		}, wantStatus: http.StatusAccepted, wantPanic: http.ErrAbortHandler, wantPanicLog: true},
		{name: "panic after body aborts without appending JSON", handler: func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("partial response"))
			panic("private panic detail")
		}, wantStatus: http.StatusOK, wantBody: "partial response", wantPanic: http.ErrAbortHandler, wantPanicLog: true},
		{name: "abort propagates without a panic log", handler: func(http.ResponseWriter, *http.Request) {
			panic(http.ErrAbortHandler)
		}, wantStatus: http.StatusOK, wantPanic: http.ErrAbortHandler},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var logs bytes.Buffer
			log := slog.New(slog.NewJSONHandler(&logs, nil))
			handler := middleware.LogAndRecover(log)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Test", "preserved")
				tt.handler(w, r)
			}))
			rr := httptest.NewRecorder()
			const reqID = "test-request-id"
			r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/payments?secret=hidden", nil)
			r.Header.Set("X-Request-ID", reqID)
			if got := capturePanic(func() { handler.ServeHTTP(rr, r) }); got != tt.wantPanic {
				t.Fatalf("panic = %v, want %v", got, tt.wantPanic)
			}
			if rr.Code != tt.wantStatus || rr.Body.String() != tt.wantBody {
				t.Errorf("response = %d %q, want %d %q", rr.Code, rr.Body.String(), tt.wantStatus, tt.wantBody)
			}
			header := rr.Result().Header
			if got := header.Get("X-Request-ID"); got != reqID {
				t.Errorf("response X-Request-ID = %q, want %q", got, reqID)
			}
			if got := header.Get("X-Test"); got != "preserved" {
				t.Errorf("response X-Test = %q, want preserved", got)
			}

			records := bytes.Split(bytes.TrimSpace(logs.Bytes()), []byte("\n"))
			if tt.wantPanicLog {
				if len(records) != 2 {
					t.Fatalf("want a panic log and an access log; got: %s", &logs)
				}
				assertPanicLog(t, records[0], reqID)
				records = records[1:]
			}
			if len(records) != 1 {
				t.Fatalf("want one access log; got: %s", &logs)
			}
			record := decodeLog(t, records[0])
			for key, want := range map[string]any{
				"level": "INFO", "msg": "request", "method": http.MethodPost, "request_id": reqID,
				"path": "/payments", "status": float64(tt.wantStatus),
			} {
				if record[key] != want {
					t.Errorf("log %s = %v, want %v", key, record[key], want)
				}
			}
			if duration, ok := record["duration"].(float64); !ok || duration < 0 {
				t.Errorf("duration = %v, want nonnegative duration", record["duration"])
			}
		})
	}
}

func TestLogAndRecover_BeforeCommit(t *testing.T) {
	t.Parallel()

	var logs bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&logs, nil))
	handler := middleware.LogAndRecover(log)(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("private panic detail")
	}))
	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/payments?secret=hidden", nil)
	handler.ServeHTTP(w, r)

	response := w.Result()
	if response.StatusCode != http.StatusInternalServerError {
		t.Fatalf("response status = %d, want 500", response.StatusCode)
	}
	if got := response.Header.Get("Content-Type"); got != "application/problem+json" {
		t.Errorf("Content-Type = %q, want application/problem+json", got)
	}
	var body map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("decode problem response: %v", err)
	}
	want := map[string]any{
		"type": "about:blank", "title": "Internal Server Error", "status": float64(500),
		"detail": "Unexpected error has occurred", "instance": "/payments",
	}
	if !reflect.DeepEqual(body, want) {
		t.Errorf("problem response = %v, want %v", body, want)
	}
	reqID := response.Header.Get("X-Request-ID")
	if reqID == "" {
		t.Fatal("response has no X-Request-ID")
	}
	records := bytes.Split(bytes.TrimSpace(logs.Bytes()), []byte("\n"))
	if len(records) != 2 {
		t.Fatalf("want a panic log and an access log; got: %s", &logs)
	}
	assertPanicLog(t, records[0], reqID)
	accessLog := decodeLog(t, records[1])
	if accessLog["msg"] != "request" || accessLog["request_id"] != reqID || accessLog["status"] != float64(http.StatusInternalServerError) {
		t.Errorf("access log = %v; want request_id %q and status 500", accessLog, reqID)
	}
}

func assertPanicLog(t *testing.T, line []byte, reqID string) {
	t.Helper()
	panicLog := decodeLog(t, line)
	for key, want := range map[string]any{
		"level": "ERROR", "msg": "panic recovered", "value": "private panic detail", "request_id": reqID,
		"method": http.MethodPost, "path": "/payments",
	} {
		if panicLog[key] != want {
			t.Errorf("panic log %s = %v, want %v", key, panicLog[key], want)
		}
	}
	if stack, ok := panicLog["stack"].(string); !ok || stack == "" {
		t.Error("panic log has no stack")
	}
}

func capturePanic(fn func()) (value any) {
	defer func() { value = recover() }()
	fn()
	return nil
}

func decodeLog(t *testing.T, logs []byte) map[string]any {
	t.Helper()
	var record map[string]any
	if err := json.Unmarshal(logs, &record); err != nil {
		t.Fatalf("decode single log record: %v; log: %s", err, logs)
	}
	return record
}
