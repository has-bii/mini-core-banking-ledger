package middleware

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogging(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		handler    http.HandlerFunc
		wantStatus int
		wantBody   string
		wantPanic  any
	}{
		{"no writes defaults to 200", func(http.ResponseWriter, *http.Request) {}, http.StatusOK, "", nil},
		{"body implies 200", func(w http.ResponseWriter, r *http.Request) {
			_, _ = w.Write([]byte("hello"))
		}, http.StatusOK, "hello", nil},
		{"explicit status", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNoContent)
		}, http.StatusNoContent, "", nil},
		{"first final status wins", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
			w.WriteHeader(http.StatusInternalServerError)
		}, http.StatusCreated, "", nil},
		{"panic still logs and propagates", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusAccepted)
			panic("handler panic")
		}, http.StatusAccepted, "", "handler panic"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var logs bytes.Buffer
			log := slog.New(slog.NewJSONHandler(&logs, nil))
			rr := httptest.NewRecorder()
			const reqID = "test-request-id"
			r := httptest.NewRequestWithContext(t.Context(), http.MethodPost, "/payments?secret=hidden", nil)
			r.Header.Set("X-Request-ID", reqID)
			gotPanic := capturePanic(func() { Logging(log)(tt.handler).ServeHTTP(rr, r) })
			if gotPanic != tt.wantPanic {
				t.Fatalf("panic = %v, want %v", gotPanic, tt.wantPanic)
			}
			if rr.Code != tt.wantStatus || rr.Body.String() != tt.wantBody {
				t.Errorf("response = %d %q, want %d %q", rr.Code, rr.Body.String(), tt.wantStatus, tt.wantBody)
			}
			if got := rr.Result().Header.Get("X-Request-ID"); got != reqID {
				t.Errorf("response X-Request-ID = %q, want %q", got, reqID)
			}
			record := decodeLog(t, logs.Bytes())
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

func TestLoggingWithRecoverer(t *testing.T) {
	t.Parallel()

	var logs bytes.Buffer
	log := slog.New(slog.NewJSONHandler(&logs, nil))
	handler := Chain(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("handler panic")
	}), Logging(log), Recoverer())
	w := httptest.NewRecorder()
	r := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/payments", nil)
	handler.ServeHTTP(w, r)

	response := w.Result()
	if response.StatusCode != http.StatusInternalServerError {
		t.Fatalf("response status = %d, want 500", response.StatusCode)
	}
	reqID := response.Header.Get("X-Request-ID")
	if reqID == "" {
		t.Fatal("response has no X-Request-ID")
	}
	records := bytes.Split(bytes.TrimSpace(logs.Bytes()), []byte("\n"))
	if len(records) != 2 {
		t.Fatalf("want a panic log and an access log; got: %s", &logs)
	}
	panicLog, accessLog := decodeLog(t, records[0]), decodeLog(t, records[1])
	if panicLog["msg"] != "panic recovered" || panicLog["request_id"] != reqID {
		t.Errorf("panic log message = %v, request_id = %v; want panic recovered with request_id %q", panicLog["msg"], panicLog["request_id"], reqID)
	}
	if accessLog["msg"] != "request" || accessLog["request_id"] != reqID || accessLog["status"] != float64(http.StatusInternalServerError) {
		t.Errorf("access log = %v; want request_id %q and status 500", accessLog, reqID)
	}
}
