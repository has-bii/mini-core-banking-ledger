package middleware

import (
	"backend/internal/platform/logger"

	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"
)

// LogAndRecover returns middleware that logs requests and recovers panics.
// It echoes the X-Request-ID header or generates a replacement if missing or
// longer than 64 bytes. A panic produces a 500 response if no final status has
// been sent, or aborts the response otherwise. Explicit handler aborts propagate.
func LogAndRecover(log *slog.Logger) Middleware {
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

			request := r.WithContext(ctx)
			// Recover first so the access log sees the recovered status.
			defer recoverRequest(rw, request, loggerWithReqID)
			next.ServeHTTP(rw, request)
		})
	}
}

func generateID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
