package httpx

import (
	"backend/internal/platform/v2/validator"
	"encoding/json"
	"errors"
	"net/http"
)

type SuccessResponse struct {
	Data    any    `json:"data"`
	Message string `json:"message,omitempty"`
}

type ErrorResponse struct {
	Type     string                  `json:"type"`
	Title    string                  `json:"title"`
	Status   int                     `json:"status"`
	Detail   string                  `json:"detail"`
	Instance string                  `json:"instance"`
	Error    *[]validator.ErrorField `json:"error,omitempty"`
}

func NewErrorResponse(title string, status int, detail string, instance string, errors *[]validator.ErrorField) *ErrorResponse {
	return &ErrorResponse{
		Type:     "about:blank",
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: instance,
		Error:    errors,
	}
}

func NewSuccessResponse(message string, data any) *SuccessResponse {
	return &SuccessResponse{Message: message, Data: data}
}

func writeJSON(w http.ResponseWriter, status int, contentType string, v any) error {
	body, err := json.Marshal(v)
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
	return writeJSON(w, v.Status, "application/problem+json", v)
}
