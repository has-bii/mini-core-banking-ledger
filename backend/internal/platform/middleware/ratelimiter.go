package middleware

import (
	"backend/internal/platform/httpx"
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	tokens     float64
	capacity   float64
	refillRate float64
	lastRefill time.Time
	mu         sync.Mutex
}

func NewRateLimiter(capacity, refillRate float64) *RateLimiter {
	return &RateLimiter{
		tokens:     capacity,
		capacity:   capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

func (tb *RateLimiter) refill(now time.Time) {
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens += elapsed * tb.refillRate

	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}

	tb.lastRefill = now
}

func (tb *RateLimiter) allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()

	now := time.Now()
	tb.refill(now)

	if tb.tokens >= 1 {
		tb.tokens--
		return true
	}

	return false
}

func (tb *RateLimiter) Middleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !tb.allow() {
				_ = httpx.SendError(w, httpx.NewErrorResponse(
					http.StatusText(http.StatusTooManyRequests),
					http.StatusTooManyRequests,
					"Rate limit exceeded",
					r.URL.Path,
					nil,
				))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
