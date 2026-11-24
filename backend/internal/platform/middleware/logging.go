package middleware

import (
	"backend/internal/platform/logger"

	"crypto/rand"
	"encoding/hex"
	"log/slog"
	"net/http"
	"time"
)

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
