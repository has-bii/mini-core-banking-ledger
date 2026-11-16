package response

import (
	"backend/internal/platform/v2/validator"
	"encoding/json"
	"net/http"
)

type SuccessResponseDTO struct {
	Data    any    `json:"data"`
	Message string `json:"message,omitempty"`
}

type ErrorResponseDto struct {
	Type     string                     `json:"type"`
	Title    string                     `json:"title"`
	Status   int                        `json:"status"`
	Detail   string                     `json:"detail"`
	Instance string                     `json:"instance"`
	Error    *validator.ErrorValidation `json:"error,omitempty"`
}

func NewErrorResponse(title string, status int, detail string, instance string, errors *validator.ErrorValidation) *ErrorResponseDto {
	return &ErrorResponseDto{
		Type:     "about:blank",
		Title:    title,
		Status:   status,
		Detail:   detail,
		Instance: instance,
		Error:    errors,
	}
}

func SendSuccess(w http.ResponseWriter, status int, v *SuccessResponseDTO) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	_ = json.NewEncoder(w).Encode(v)
}

func SendError(w http.ResponseWriter, status int, v *ErrorResponseDto) {
	w.Header().Set("Content-Type", "application/problem+json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
