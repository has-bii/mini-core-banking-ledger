package httpx

import (
	"net/http"

	"golang.org/x/time/rate"
)

func RateLimit(limiter *rate.Limiter) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.Allow() {
				w.Header().Set("Cache-Control", "no-store")
				WriteError(w, http.StatusTooManyRequests, CodeRateLimited, "too many requests")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
