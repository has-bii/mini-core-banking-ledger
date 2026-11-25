package middleware

import (
	"backend/internal/platform/httpx"

	"log/slog"
	"net/http"
	"runtime/debug"
)

// recoverRequest must be deferred. It logs unexpected panics and sends a 500
// before commitment or aborts an already committed response.
func recoverRequest(w *responseWriter, r *http.Request, log *slog.Logger) {
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

	if w.committed {
		panic(http.ErrAbortHandler)
	}

	w.Header().Del("Set-Cookie")
	httpx.SendError(w, r, httpx.NewHTTPError(http.StatusInternalServerError, "Unexpected error has occurred"))
}
