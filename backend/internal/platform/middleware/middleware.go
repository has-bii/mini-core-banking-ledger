package middleware

import (
	"backend/internal/platform/httpx"
	"backend/internal/platform/logger"
	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"runtime/debug"
	"slices"
	"time"
)

type Middleware func(http.Handler) http.Handler

func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for _, mw := range slices.Backward(mws) {
		h = mw(h)
	}
	return h
}

func Recoverer() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}

			log := logger.FromContext(r.Context())

			defer func() {
				v := recover()
				if v == nil {
					return
				}
				if v == http.ErrAbortHandler {
					panic(v)
				}

				log.ErrorContext(r.Context(), "panic recovered",
					slog.Any("value", v),
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("stack", string(debug.Stack())),
				)

				if rw.committed {
					panic(http.ErrAbortHandler)
				}

				_ = httpx.SendError(rw, http.StatusInternalServerError, "Unexpected error has occurred", r.URL.Path, nil)
			}()
			next.ServeHTTP(rw, r)
		})
	}
}

func Logging(log *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			rw := &responseWriter{
				ResponseWriter: w,
				status:         http.StatusOK,
			}

			reqID := r.Header.Get("X-Request-ID")
			if reqID == "" || len(reqID) > 64 {
				reqID = generateID()
			}
			rw.Header().Set("X-Request-ID", reqID)

			loggerWithReqID := log.With("request_id", reqID)

			ctx := logger.WithContext(r.Context(), loggerWithReqID)

			defer func() {
				loggerWithReqID.InfoContext(ctx, "request",
					"method", r.Method,
					"path", r.URL.Path,
					"status", rw.status,
					"duration", time.Since(start),
				)
			}()

			next.ServeHTTP(rw, r.WithContext(ctx))
		})
	}
}

func generateID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

/* ----------------------------- Response Writer ---------------------------- */

type responseWriter struct {
	http.ResponseWriter
	committed bool
	status    int
}

func (w *responseWriter) WriteHeader(status int) {
	if w.committed {
		return
	}

	if status == http.StatusSwitchingProtocols || status >= 200 {
		w.committed = true
		w.status = status
	}

	w.ResponseWriter.WriteHeader(status)
}

func (w *responseWriter) Write(p []byte) (int, error) {
	if !w.committed {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(p)
}

func (w *responseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
