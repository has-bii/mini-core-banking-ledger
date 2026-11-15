package middleware_test

import (
	"backend/internal/platform/v2/middleware"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRateLimiter_Middleware(t *testing.T) {
	t.Parallel()

	calls := 0
	handler := middleware.NewRateLimiter(1, 0).Middleware()(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusNoContent)
	}))

	for _, want := range []int{http.StatusNoContent, http.StatusTooManyRequests, http.StatusTooManyRequests} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/", nil))
		if response.Code != want {
			t.Fatalf("status = %d, want %d", response.Code, want)
		}
		if want == http.StatusTooManyRequests && !strings.Contains(response.Body.String(), `{"type":"about:blank","title":"Too Many Requests","status":429,"detail":"Rate limit exceeded","instance":"/"}`) {
			t.Fatalf("unexpected error response: %s", response.Body.String())
		}
	}
	if calls != 1 {
		t.Fatalf("handler called %d times, want 1", calls)
	}
}
