package middleware

import (
	"backend/internal/platform/logger"

	"bytes"
	"encoding/json/v2"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

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
			log := slog.New(slog.NewJSONHandler(&logs, nil))
			ctx := logger.WithContext(t.Context(), log)
			rr := httptest.NewRecorder()
			r := httptest.NewRequestWithContext(ctx, http.MethodPost, "/payments?secret=hidden", nil)
			gotPanic := capturePanic(func() { Recoverer()(tt.handler).ServeHTTP(rr, r) })
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
			record := decodeLog(t, logs.Bytes())
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
