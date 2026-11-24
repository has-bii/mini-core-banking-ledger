package httpx

import (
	"backend/internal/platform/logger"
	"backend/internal/platform/validator"

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
		// A replacement response must not send a staged success cookie or depend on encoding again.
		w.Header().Del("Set-Cookie")
		status = http.StatusInternalServerError
		contentType = problemJSON
		encoded = []byte(`{"type":"about:blank","title":"Internal Server Error","status":500,"detail":"internal error","instance":""}`)
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

func SendError(w http.ResponseWriter, r *http.Request, status int, detail string, fields *validator.ErrorValidation) {
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

// SendRequestError handles transport errors before consulting mapDomain, which returns status 0 for unrecognized errors.
func SendRequestError(w http.ResponseWriter, r *http.Request, scope string, err error, mapDomain func(error) (int, string)) {
	status, detail := http.StatusInternalServerError, "internal error"
	var httpErr *HTTPError
	var fields *validator.ErrorValidation
	switch {
	case errors.As(err, &httpErr):
		status, detail = httpErr.Status, httpErr.Message
	case errors.As(err, &fields):
		status, detail = http.StatusBadRequest, fields.Error()
	default:
		if mappedStatus, mappedDetail := mapDomain(err); mappedStatus != 0 {
			status, detail = mappedStatus, mappedDetail
		}
	}
	if status == http.StatusInternalServerError {
		ctx := r.Context()
		logger.FromContext(ctx).ErrorContext(ctx, scope+" request failed", "err", err, "method", r.Method, "path", r.URL.Path)
	}
	SendError(w, r, status, detail, fields)
}
