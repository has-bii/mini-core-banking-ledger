package ratelimit

import (
	"backend/internal/platform/httpx"
	"net/http"
	"sync"
	"time"
)

type TokenBucket struct {
	tokens     float64
	capacity   float64
	refillRate float64
	lastRefill time.Time
	mu         sync.Mutex
}

func New(capacity, refillRate float64) *TokenBucket {
	return &TokenBucket{
		tokens:     capacity,
		capacity:   capacity,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

func (tb *TokenBucket) refill(now time.Time) {
	elapsed := now.Sub(tb.lastRefill).Seconds()
	tb.tokens += elapsed * tb.refillRate

	if tb.tokens > tb.capacity {
		tb.tokens = tb.capacity
	}

	tb.lastRefill = now
}

func (tb *TokenBucket) allow() bool {
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

func (tb *TokenBucket) Middleware() httpx.Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !tb.allow() {
				httpx.WriteError(w, http.StatusTooManyRequests, httpx.CodeTooManyRequests, "too many requests")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
