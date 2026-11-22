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

func writeJSON(w http.ResponseWriter, status int, contentType string, v any) error {
	body, err := json.Marshal(v, json.FormatNilMapAsNull(true), json.FormatNilSliceAsNull(true))
	if err != nil {
		return err
	}

	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)

	_, err = w.Write(append(body, '\n'))
	return err
}

func SendSuccess(w http.ResponseWriter, status int, response SuccessResponse) error {
	return writeJSON(w, status, "application/json", response)
}

func SendError(w http.ResponseWriter, status int, detail, instance string, fields *validator.ErrorValidation) error {
	response := ErrorResponse{
		Type:     "about:blank",
		Title:    http.StatusText(status),
		Status:   status,
		Detail:   detail,
		Instance: strings.ToValidUTF8(instance, "\uFFFD"), // A decoded request path can contain invalid UTF-8.
	}
	if fields != nil {
		response.ErrorValidation = *fields
	}
	return writeJSON(w, status, "application/problem+json", response)
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
	ctx := r.Context()
	if status == http.StatusInternalServerError {
		logger.FromContext(ctx).ErrorContext(ctx, scope+" request failed", "err", err, "method", r.Method, "path", r.URL.Path)
	}
	if err := SendError(w, status, detail, r.URL.Path, fields); err != nil {
		logger.FromContext(ctx).ErrorContext(ctx, "write "+scope+" error response", "err", err)
	}
}
