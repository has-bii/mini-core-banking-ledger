package middleware

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestChain(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name  string
		count int
	}{
		{"no middleware", 0},
		{"one middleware", 1},
		{"multiple middleware", 3},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			count := tt.count
			var got, want []string
			var mws []Middleware
			for i := range count {
				name := string(rune('A' + i))
				want = append(want, name+" before")
				mws = append(mws, func(next http.Handler) http.Handler {
					return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
						got = append(got, name+" before")
						next.ServeHTTP(w, r)
						got = append(got, name+" after")
					})
				})
			}
			want = append(want, "handler")
			for i := count - 1; i >= 0; i-- {
				want = append(want, string(rune('A'+i))+" after")
			}
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				got = append(got, "handler")
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte("created"))
			})
			rr := httptest.NewRecorder()
			Chain(handler, mws...).ServeHTTP(rr, httptest.NewRequest(http.MethodGet, "/", nil))
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("execution order = %v, want %v", got, want)
			}
			if rr.Code != http.StatusCreated || rr.Body.String() != "created" {
				t.Fatalf("response = %d %q, want 201 created", rr.Code, rr.Body.String())
			}
		})
	}
}
