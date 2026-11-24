package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

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
