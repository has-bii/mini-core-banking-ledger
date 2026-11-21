package middleware

import (
	"bytes"
	"encoding/json/v2"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func TestChain(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name  string
		count int
	}{
		{"no middleware", 0},
		{"one middleware", 1},
		{"multiple middleware", 3},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			count := tt.count
			var got, want []string
			var mws []Middleware
			for i := range count {
				name := string(rune('A' + i))
				want = append(want, name+" before")
				mws = append(mws, func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						got = append(got, name+" before")
						next.ServeHTTP(w, r)
						got = append(got, name+" after")
					})
				})
			}
			want = append(want, "handler")
			for i := count - 1; i >= 0; i-- {
				want = append(want, string(rune('A'+i))+" after")
			}
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got = append(got, "handler")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte("created"))
			})
			rr := httptest.NewRecorder()
			Chain(handler, mws...).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("execution order = %v, want %v", got, want)
			}
			if rr.Code != http.StatusCreated || rr.Body.String() != "created" {
				t.Fatalf("response = %d %q, want 201 created", rr.Code, rr.Body.String())
			}
		})
	}
}

func TestRecoverer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		handler    http.HandlerFunc
		wantStatus int
		wantBody   string
		wantPanic  any
		wantLog    bool
		wantJSON   bool
	}{
		{
			name: "normal response passes through",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("X-Test", "preserved")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte("created"))
			},
			wantStatus: http.StatusCreated,
			wantBody:   "created",
		},
		{
			name:       "panic before response returns problem JSON",
			handler:    func(http.ResponseWriter, *http.Request) { panic("private panic detail") },
			wantStatus: http.StatusInternalServerError,
			wantLog:    true,
			wantJSON:   true,
		},
		{
			name:       "abort handler propagates without logging",
			handler:    func(http.ResponseWriter, *http.Request) { panic(http.ErrAbortHandler) },
			wantStatus: http.StatusOK,
			wantPanic:  http.ErrAbortHandler,
		},
		{
			name: "panic after final headers aborts response",
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusAccepted)
				panic("private panic detail")
			},
			wantStatus: http.StatusAccepted,
			wantPanic:  http.ErrAbortHandler,
			wantLog:    true,
		},
		{
			name: "panic after body aborts without appending JSON",
			handler: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("partial response"))
				panic("private panic detail")
			},
			wantStatus: http.StatusOK,
			wantBody:   "partial response",
			wantPanic:  http.ErrAbortHandler,
			wantLog:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var logs bytes.Buffer
			logger := slog.New(slog.NewJSONHandler(&logs, nil))
			rr := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/payments?secret=hidden", nil)
			gotPanic := capturePanic(func() { Recoverer(logger)(tt.handler).ServeHTTP(rr, r) })
			if gotPanic != tt.wantPanic {
				t.Fatalf("panic = %v, want %v", gotPanic, tt.wantPanic)
			}
			if rr.Code != tt.wantStatus {
				t.Fatalf("status = %d, want %d", rr.Code, tt.wantStatus)
			}
			if tt.wantJSON {
				if got := rr.Header().Get("Content-Type"); got != "application/problem+json" {
					t.Errorf("Content-Type = %q, want application/problem+json", got)
				}
				var body map[string]any
				if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
					t.Fatalf("decode error response: %v", err)
				}
				want := map[string]any{
					"type": "about:blank", "title": "Internal Server Error", "status": float64(500),
					"detail": "Unexpected error has occurred", "instance": "/payments",
				}
				if !reflect.DeepEqual(body, want) {
					t.Errorf("problem response = %v, want %v", body, want)
				}
				if strings.Contains(rr.Body.String(), "private panic detail") {
					t.Error("response leaks panic detail")
				}
			} else if rr.Body.String() != tt.wantBody {
				t.Errorf("body = %q, want %q", rr.Body.String(), tt.wantBody)
			}
			if tt.name == "normal response passes through" && rr.Header().Get("X-Test") != "preserved" {
				t.Error("handler header was not preserved")
			}
			if !tt.wantLog {
				if logs.Len() != 0 {
					t.Errorf("unexpected log: %s", &logs)
				}
				return
			}
			record := decodeLog(t, &logs)
			for key, want := range map[string]any{
				"level": "ERROR", "msg": "panic recovered", "value": "private panic detail",
				"method": http.MethodPost, "path": "/payments",
			} {
				if record[key] != want {
					t.Errorf("log %s = %v, want %v", key, record[key], want)
				}
			}
			if stack, ok := record["stack"].(string); !ok || stack == "" {
				t.Error("panic log has no stack")
			}
		})
	}
}

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
			logger := slog.New(slog.NewJSONHandler(&logs, nil))
			rr := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/payments?secret=hidden", nil)
			gotPanic := capturePanic(func() { Logging(logger)(tt.handler).ServeHTTP(rr, r) })
			if gotPanic != tt.wantPanic {
				t.Fatalf("panic = %v, want %v", gotPanic, tt.wantPanic)
			}
			if rr.Code != tt.wantStatus || rr.Body.String() != tt.wantBody {
				t.Errorf("response = %d %q, want %d %q", rr.Code, rr.Body.String(), tt.wantStatus, tt.wantBody)
			}
			record := decodeLog(t, &logs)
			for key, want := range map[string]any{
				"level": "INFO", "msg": "request", "method": http.MethodPost,
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

func TestResponseWriter_WriteHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		statuses      []int
		wantWrites    []int
		wantStatus    int
		wantCommitted bool
	}{
		{"final response", []int{201}, []int{201}, 201, true},
		{"duplicate final response ignored", []int{201, 500}, []int{201}, 201, true},
		{"informational response is not committed", []int{103}, []int{103}, 200, false},
		{"informational responses precede final response", []int{100, 103, 201}, []int{100, 103, 201}, 201, true},
		{"switching protocols commits response", []int{101, 500}, []int{101}, 101, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			underlying := &recordingWriter{header: make(http.Header)}
			w := &responseWriter{ResponseWriter: underlying, status: http.StatusOK}
			for _, status := range tt.statuses {
				w.WriteHeader(status)
			}
			if !reflect.DeepEqual(underlying.statuses, tt.wantWrites) {
				t.Errorf("forwarded statuses = %v, want %v", underlying.statuses, tt.wantWrites)
			}
			if w.status != tt.wantStatus {
				t.Errorf("tracked status = %d, want %d", w.status, tt.wantStatus)
			}
			if w.committed != tt.wantCommitted {
				t.Errorf("committed = %v, want %v", w.committed, tt.wantCommitted)
			}
		})
	}
}

func TestResponseWriter_Write(t *testing.T) {
	t.Parallel()

	writeErr := errors.New("write failed")
	tests := []struct {
		name       string
		status     int
		wantStatus int
		wantN      int
		wantErr    error
	}{
		{"implicit status", 0, http.StatusOK, 5, nil},
		{"explicit status preserved", http.StatusCreated, http.StatusCreated, 5, nil},
		{"informational response followed by body", http.StatusEarlyHints, http.StatusOK, 5, nil},
		{"partial write error returned", 0, http.StatusOK, 2, writeErr},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			underlying := &recordingWriter{header: make(http.Header), n: tt.wantN, err: tt.wantErr}
			w := &responseWriter{ResponseWriter: underlying, status: http.StatusOK}
			var wantStatuses []int
			if tt.status != 0 {
				w.WriteHeader(tt.status)
				wantStatuses = append(wantStatuses, tt.status)
			}
			if tt.status == 0 || tt.status == http.StatusEarlyHints {
				wantStatuses = append(wantStatuses, http.StatusOK)
			}
			n, err := w.Write([]byte("hello"))
			if n != tt.wantN || !errors.Is(err, tt.wantErr) {
				t.Errorf("Write = (%d, %v), want (%d, %v)", n, err, tt.wantN, tt.wantErr)
			}
			if string(underlying.body) != "hello" {
				t.Errorf("forwarded body = %q, want hello", underlying.body)
			}
			if !reflect.DeepEqual(underlying.statuses, wantStatuses) {
				t.Errorf("forwarded statuses = %v, want %v", underlying.statuses, wantStatuses)
			}
			if !w.committed || w.status != tt.wantStatus {
				t.Errorf("writer state = (%v, %d), want (true, %d)", w.committed, w.status, tt.wantStatus)
			}
		})
	}
}

func TestResponseWriter_Unwrap(t *testing.T) {
	t.Parallel()
	underlying := httptest.NewRecorder()
	w := &responseWriter{ResponseWriter: underlying, status: http.StatusOK}
	if w.Unwrap() != underlying {
		t.Fatal("Unwrap did not return the original writer")
	}
	if err := http.NewResponseController(w).Flush(); err != nil {
		t.Fatalf("ResponseController.Flush: %v", err)
	}
	if !underlying.Flushed {
		t.Error("ResponseController did not reach the underlying flusher")
	}
}

func capturePanic(fn func()) (value any) {
	defer func() { value = recover() }()
	fn()
	return nil
}

func decodeLog(t *testing.T, logs *bytes.Buffer) map[string]any {
	t.Helper()
	var record map[string]any
	if err := json.Unmarshal(logs.Bytes(), &record); err != nil {
		t.Fatalf("decode single log record: %v; log: %s", err, logs)
	}
	return record
}

type recordingWriter struct {
	header   http.Header
	statuses []int
	body     []byte
	n        int
	err      error
}

func (w *recordingWriter) Header() http.Header { return w.header }

func (w *recordingWriter) WriteHeader(status int) {
	w.statuses = append(w.statuses, status)
}

func (w *recordingWriter) Write(p []byte) (int, error) {
	w.body = append(w.body, p...)
	return w.n, w.err
}
