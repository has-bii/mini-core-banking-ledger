package httpx_test

import (
	"backend/internal/platform/httpx"
	"errors"
	"net/http"
	"testing"
)

func TestWrapHTTPErrorPreservesCause(t *testing.T) {
	t.Parallel()
	cause := errors.New("backend unavailable")
	err := httpx.WrapHTTPError(http.StatusServiceUnavailable, "please retry later", cause)
	if got := errors.Unwrap(err); got != cause {
		t.Errorf("unwrapped error = %v, want original cause %v", got, cause)
	}
}
