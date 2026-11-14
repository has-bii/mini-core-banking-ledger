package response

import (
	"backend/internal/platform/v2/validator"
	"encoding/json"
	"net/http"
)

type SuccessResponseData struct {
	Data    any    `json:"data"`
	Message string `json:"message,omitempty"`
}

type ErrorResponseData struct {
	Type     string                     `json:"type"`
	Title    string                     `json:"title"`
	Status   int                        `json:"status"`
	Detail   string                     `json:"detail"`
	Instance string                     `json:"instance"`
	Error    *validator.ErrorValidation `json:"error,omitempty"`
}

func NewErrorResponseData(title string, status int, detail string, instance string, errors *validator.ErrorValidation) *ErrorResponseData {
	return &ErrorResponseData{
		Type:     "about:blank",
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: instance,
		Error:    errors,
	}
}

func HttpResponse(w http.ResponseWriter, status int, v any) {
	contentType := "application/json"
	if status >= 400 {
		contentType = "application/problem+json"
	}

	w.Header().Set("Content-Type", contentType)

	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(v)
}

func SuccessResponse(w http.ResponseWriter, status int, v *SuccessResponseData) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(v)
}

func ErrorResponse(w http.ResponseWriter, status int, v *ErrorResponseData) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
