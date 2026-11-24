package middleware

import (
	"backend/internal/platform/httpx"
	"backend/internal/platform/logger"

	"log/slog"
	"net/http"
	"runtime/debug"
)

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

				httpx.SendError(rw, r, http.StatusInternalServerError, "Unexpected error has occurred", nil)
			}()
			next.ServeHTTP(rw, r)
		})
	}
}
