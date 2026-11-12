package httpx

import (
	"backend/internal/platform/apperror"
	"backend/internal/platform/validation"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

var (
	ErrSingleJSONValueRequired = errors.New("request body must contain one JSON value")
	errBodyRequired            = errors.New("request body is required")
	errJSONObjectRequired      = errors.New("request body must be a JSON object")
)

func WriteJSON(w http.ResponseWriter, status int, v any) {
	writeJSON(w, status, map[string]any{"data": v})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	// Headers are committed; a write failure cannot be replaced with another response.
	_ = json.NewEncoder(w).Encode(v)
}

func WriteError(w http.ResponseWriter, status int, code, msg string, details ...validation.FieldError) {
	var fields any
	if len(details) > 0 {
		fields = details
	}
	writeJSON(w, status, map[string]any{"error": apperror.New(code, msg, fields, nil)})
}

func WriteValidationError(w http.ResponseWriter, errs validation.Errors) {
	WriteError(w, http.StatusBadRequest, apperror.CodeValidationError, "Some fields are invalid.", errs...)
}

func ValidationError(errs validation.Errors) error {
	return apperror.New(apperror.CodeValidationError, "Some fields are invalid.", errs, errs)
}

func WriteDecodeError(w http.ResponseWriter, err error) {
	appErr := DecodeError(err)
	writeJSON(w, errorStatus(appErr.Code), map[string]any{"error": appErr})
}

func DecodeError(err error) *apperror.Error {
	var typeErr *json.UnmarshalTypeError
	if _, ok := errors.AsType[*http.MaxBytesError](err); ok {
		return apperror.New(apperror.CodePayloadTooLarge, "Request body must not exceed 1 MiB.", nil, err)
	}
	if errors.As(err, &typeErr) && typeErr.Field != "" {
		message := fmt.Sprintf("Field %q has an invalid type.", typeErr.Field)
		if typeErr.Type.Kind() == reflect.String {
			message = fmt.Sprintf("Field %q must be a string.", typeErr.Field)
		}
		return apperror.New(apperror.CodeInvalidJSON, message, []validation.FieldError{{
			Field: typeErr.Field, Code: validation.CodeInvalidType, Message: message,
		}}, err)
	}
	// encoding/json exposes unknown fields only through its quoted error message.
	if quoted, ok := strings.CutPrefix(err.Error(), "json: unknown field "); ok {
		if field, unquoteErr := strconv.Unquote(quoted); unquoteErr == nil {
			message := fmt.Sprintf("Field %q is not allowed.", field)
			return apperror.New(apperror.CodeInvalidJSON, message, []validation.FieldError{{
				Field: field, Code: validation.CodeUnknownField, Message: message,
			}}, err)
		}
	}
	message := "Request body contains malformed JSON."
	switch {
	case errors.Is(err, errBodyRequired):
		message = "Request body is required."
	case errors.Is(err, errJSONObjectRequired), typeErr != nil:
		message = "Request body must be a JSON object."
	case errors.Is(err, ErrSingleJSONValueRequired):
		message = "Request body must contain one JSON object."
	}
	return apperror.New(apperror.CodeInvalidJSON, message, nil, err)
}

func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20)
	// Read the bounded body first so trailing data cannot bypass the size limit.
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return err
	}
	body = bytes.Trim(body, " \t\r\n")
	if len(body) == 0 {
		return errBodyRequired
	}
	if body[0] != '{' && json.Valid(body) {
		return errJSONObjectRequired
	}
	dec := json.NewDecoder(bytes.NewReader(body))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if err := dec.Decode(new(any)); err != io.EOF {
		if err != nil {
			return err
		}
		return ErrSingleJSONValueRequired
	}
	return nil
}
