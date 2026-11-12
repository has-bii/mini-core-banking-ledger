package httpx

import (
	"backend/internal/platform/apperror"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
)

// Adapt handles returned errors centrally. Handlers must return errors before writing a response.
func Adapt(h func(http.ResponseWriter, *http.Request) error, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := h(w, r); err != nil {
			HandleError(w, r, err, log)
		}
	}
}

// HandleError also serves middleware, whose standard HTTP handlers cannot return errors.
func HandleError(w http.ResponseWriter, r *http.Request, err error, log *slog.Logger) {
	var appErr *apperror.Error
	status := http.StatusInternalServerError
	if errors.As(err, &appErr) && appErr != nil {
		status = errorStatus(appErr.Code)
	}
	if status != http.StatusInternalServerError {
		body, encodeErr := json.Marshal(map[string]any{"error": appErr})
		if encodeErr == nil {
			writeJSON(w, status, json.RawMessage(body))
			return
		}
		err = errors.Join(err, encodeErr)
	}
	if log == nil {
		log = slog.Default()
	}
	var cause error
	if appErr != nil {
		cause = appErr.Unwrap()
	}
	log.ErrorContext(r.Context(), "request failed", "err", err, "cause", cause, "method", r.Method, "path", r.URL.Path)
	WriteError(w, http.StatusInternalServerError, apperror.CodeInternalError, "internal error")
}

func errorStatus(code string) int {
	switch code {
	case apperror.CodeNotFound:
		return http.StatusNotFound
	case apperror.CodeInvalidJSON, apperror.CodeValidationError:
		return http.StatusBadRequest
	case apperror.CodePayloadTooLarge:
		return http.StatusRequestEntityTooLarge
	case apperror.CodeUnauthenticated, apperror.CodeInvalidCredentials:
		return http.StatusUnauthorized
	case apperror.CodeForbidden:
		return http.StatusForbidden
	case apperror.CodeEmailTaken:
		return http.StatusConflict
	case apperror.CodeTooManyRequests:
		return http.StatusTooManyRequests
	case apperror.CodeUnprocessableEntity:
		return http.StatusUnprocessableEntity
	default:
		return http.StatusInternalServerError
	}
}
