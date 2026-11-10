package httpx_test

import (
	"backend/internal/platform/transport/httpx"
	"net/http/httptest"
	"testing"
)

func TestNoStoreHeader(t *testing.T) {
	t.Parallel()
	for _, previous := range []string{"", "public, max-age=3600"} {
		t.Run(previous, func(t *testing.T) {
			t.Parallel()
			w := httptest.NewRecorder()
			if previous != "" {
				w.Header().Set("Cache-Control", previous)
			}

			httpx.NoStoreHeader(w)

			values := w.Header().Values("Cache-Control")
			if len(values) != 1 || values[0] != "no-store" {
				t.Errorf("Cache-Control = %v, want [no-store]", values)
			}
		})
	}
}
