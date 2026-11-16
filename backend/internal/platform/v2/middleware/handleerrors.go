package middleware

import (
	"backend/internal/platform/v2/httpx"
	"backend/internal/platform/v2/validator"
	"errors"
	"log/slog"
	"net/http"
)

func handleError(w http.ResponseWriter, r *http.Request, err error, log *slog.Logger) {
	logError := func(message string, err error) {
		log.ErrorContext(r.Context(), message,
			"error", err, "method", r.Method, "path", r.URL.Path)
	}

	if rw, ok := w.(*responseWriter); ok && rw.committed {
		logError("handler error after response committed", err)
		return
	}

	status := http.StatusInternalServerError
	detail := "Unexpected error has occurred"
	var fields *[]validator.ErrorField
	var bodyErr *httpx.MalformedRequestBody
	var validationErr *validator.ErrorValidation

	switch {
	case errors.As(err, &bodyErr):
		status, detail = bodyErr.Status, bodyErr.Msg
	case errors.As(err, &validationErr):
		status, detail, fields = http.StatusUnprocessableEntity, validationErr.Error(), validationErr.Errors
	default:
		logError("handler error", err)
	}

	if err := httpx.SendError(w, httpx.NewErrorResponse(
		http.StatusText(status), status, detail, r.URL.Path, fields,
	)); err != nil {
		logError("failed to write error response", err)
	}
}

func HandleErrors(h func(http.ResponseWriter, *http.Request) error, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rw, ok := w.(*responseWriter)
		if !ok {
			rw = &responseWriter{ResponseWriter: w, status: http.StatusOK}
		}
		if err := h(rw, r); err != nil {
			handleError(rw, r, err, log)
		}
	}
}
