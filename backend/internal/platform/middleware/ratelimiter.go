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

// NewRateLimiter returns a full token bucket with the given capacity and a
// refill rate in tokens per second.
func NewRateLimiter(capacity, refillRate float64) *RateLimiter {
	return &RateLimiter{
		tokens:     capacity,
		capacity:   capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

// refill adds tokens up to capacity. The caller must hold tb.mu.
func (tb *RateLimiter) refill(now time.Time) {
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens += elapsed * tb.refillRate

	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}

	tb.lastRefill = now
}

// allow consumes one token if available and reports whether it did so.
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

// Middleware returns middleware that shares the limiter across requests and
// responds with 429 Too Many Requests when no token is available.
func (tb *RateLimiter) Middleware() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !tb.allow() {
				httpx.SendError(w, r, httpx.NewHTTPError(http.StatusTooManyRequests, "Rate limit exceeded"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
