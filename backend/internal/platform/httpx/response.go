package httpx

import (
	"backend/internal/platform/validator"

	"encoding/json/v2"
	"errors"
	"log/slog"
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

func NewErrorResponse(title string, status int, detail string, instance string, errors *validator.ErrorValidation) *ErrorResponse {
	response := &ErrorResponse{
		Type:     "about:blank",
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: instance,
	}
	if errors != nil {
		response.ErrorValidation = *errors
	}
	return response
}

func NewSuccessResponse(message string, data any) *SuccessResponse {
	return &SuccessResponse{Message: message, Data: data}
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

func SendSuccess(w http.ResponseWriter, status int, v *SuccessResponse) error {
	if v == nil {
		return errors.New("success response must not be nil")
	}
	return writeJSON(w, status, "application/json", v)
}

func SendError(w http.ResponseWriter, v *ErrorResponse) error {
	if v == nil {
		return errors.New("error response must not be nil")
	}
	// A decoded request path can contain invalid UTF-8.
	response := *v
	response.Instance = strings.ToValidUTF8(response.Instance, "\uFFFD")
	return writeJSON(w, response.Status, "application/problem+json", &response)
}

// SendRequestError handles transport errors before consulting mapDomain, which returns status 0 for unrecognized errors.
func SendRequestError(w http.ResponseWriter, r *http.Request, log *slog.Logger, scope string, err error, mapDomain func(error) (int, string)) {
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
		log.ErrorContext(r.Context(), scope+" request failed", "err", err, "method", r.Method, "path", r.URL.Path)
	}
	if err := SendError(w, NewErrorResponse(http.StatusText(status), status, detail, r.URL.Path, fields)); err != nil {
		log.ErrorContext(r.Context(), "write "+scope+" error response", "err", err)
	}
}
