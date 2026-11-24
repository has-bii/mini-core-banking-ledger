package httpx

import (
	"backend/internal/platform/logger"
	"backend/internal/platform/validator"

	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"

	"net/http"
	"strings"
)

type SuccessResponse struct {
	Data    any    `json:"data"`
	Message string `json:"message,omitempty"`
}

type ErrorResponse struct {
	Type     string `json:"type"`
	Title    string `json:"title"`
	Status   int    `json:"status"`
	Detail   string `json:"detail"`
	Instance string `json:"instance"`
	validator.ErrorValidation
}

const problemJSON = "application/problem+json"

func writeJSON(w http.ResponseWriter, r *http.Request, status int, contentType string, body any) {
	ctx := r.Context()
	log := logger.FromContext(ctx)
	encoded, err := json.Marshal(body, json.FormatNilMapAsNull(true), json.FormatNilSliceAsNull(true))
	if err != nil {
		log.ErrorContext(ctx, "encode JSON response", "err", err, "status", status, "method", r.Method, "path", r.URL.Path)
		// A replacement response must not send a staged success cookie.
		w.Header().Del("Set-Cookie")
		status = http.StatusInternalServerError
		contentType = problemJSON
		// Quote only the path; AppendQuote cannot fail after UTF-8 normalization.
		encoded, _ = jsontext.AppendQuote([]byte(`{"type":"about:blank","title":"Internal Server Error","status":500,"detail":"internal error","instance":`), strings.ToValidUTF8(r.URL.Path, "\uFFFD"))
		encoded = append(encoded, '}')
	}

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)
	if _, err := w.Write(append(encoded, '\n')); err != nil {
		log.ErrorContext(ctx, "write JSON response", "err", err, "status", status, "method", r.Method, "path", r.URL.Path)
	}
}

func SendSuccess(w http.ResponseWriter, r *http.Request, status int, response SuccessResponse) {
	writeJSON(w, r, status, "application/json", response)
}

// SendError classifies an error and writes its client-facing problem response.
func SendError(w http.ResponseWriter, r *http.Request, err error) {
	status, detail := http.StatusInternalServerError, "internal error"
	cause := err
	var fields *validator.ErrorValidation
	if httpErr, ok := errors.AsType[*HTTPError](err); ok {
		status, detail, cause = httpErr.Status, httpErr.Message, httpErr.Unwrap()
	} else if validationErr, ok := errors.AsType[*validator.ErrorValidation](err); ok {
		status, detail, fields = http.StatusBadRequest, validationErr.Error(), validationErr
	}
	if status >= 500 && status < 600 && cause != nil {
		ctx := r.Context()
		logger.FromContext(ctx).ErrorContext(ctx, "request failed", "err", cause, "method", r.Method, "path", r.URL.Path)
	}
	response := ErrorResponse{
		Type:     "about:blank",
		Title:    http.StatusText(status),
		Status:   status,
		Detail:   detail,
		Instance: strings.ToValidUTF8(r.URL.Path, "\uFFFD"), // A decoded request path can contain invalid UTF-8.
	}
	if fields != nil {
		response.ErrorValidation = *fields
	}
	writeJSON(w, r, status, problemJSON, response)
}
