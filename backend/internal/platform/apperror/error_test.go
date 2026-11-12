package apperror_test

import (
	"backend/internal/platform/apperror"
	"encoding/json"
	"errors"
	"fmt"
	"testing"
)

func TestError_OmitsAbsentDetails(t *testing.T) {
	t.Parallel()
	err := apperror.New("NOT_FOUND", "account not found", nil, nil)
	body, encodeErr := json.Marshal(err)
	if encodeErr != nil {
		t.Fatal(encodeErr)
	}
	if string(body) != `{"code":"NOT_FOUND","message":"account not found"}` || errors.Unwrap(err) != nil {
		t.Errorf("error = %s, cause = %v; want omitted details and no cause", body, errors.Unwrap(err))
	}
}

func TestError_PreservesCauseAndPublicFields(t *testing.T) {
	t.Parallel()
	cause := errors.New("private database failure")
	appErr := apperror.New("NOT_FOUND", "account not found", map[string]any{"resource": "account"}, cause)
	wrapped := fmt.Errorf("lookup: %w", appErr)
	var got *apperror.Error
	if !errors.As(wrapped, &got) || got != appErr || !errors.Is(wrapped, cause) {
		t.Fatalf("error chain lost application error or cause: %v", wrapped)
	}
	if got.Error() != "account not found" {
		t.Errorf("Error() = %q", got.Error())
	}
	body, err := json.Marshal(got)
	if err != nil {
		t.Fatal(err)
	}
	if string(body) != `{"code":"NOT_FOUND","message":"account not found","details":{"resource":"account"}}` {
		t.Errorf("public error = %s", body)
	}
}
