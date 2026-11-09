package httpx_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"backend/internal/platform/httpx"

	"golang.org/x/time/rate"
)

func TestRateLimitBurstAndRefill(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		calls := 0
		handler := httpx.RateLimit(rate.NewLimiter(1, 2))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			calls++
			w.WriteHeader(http.StatusNoContent)
		}))
		for _, want := range []int{204, 204, 429, 429} {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
			if rec.Code != want {
				t.Fatalf("status = %d, want %d", rec.Code, want)
			}
			if want == 429 {
				var body struct {
					Error struct{ Code, Message string }
				}
				if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
					t.Fatal(err)
				}
				if body.Error.Code != httpx.CodeRateLimited || body.Error.Message != "too many requests" {
					t.Fatalf("unexpected error: %+v", body.Error)
				}
				if rec.Header().Get("Content-Type") != "application/json" || rec.Header().Get("Cache-Control") != "no-store" {
					t.Fatalf("unexpected headers: %v", rec.Header())
				}
			}
		}
		if calls != 2 {
			t.Fatalf("downstream calls = %d, want 2", calls)
		}

		time.Sleep(time.Second)
		for _, want := range []int{204, 429} {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
			if rec.Code != want {
				t.Fatalf("after refill: status = %d, want %d", rec.Code, want)
			}
		}
	})
}

func TestRateLimitConcurrentRequests(t *testing.T) {
	handler := httpx.RateLimit(rate.NewLimiter(rate.Every(time.Hour), 10))(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	var wg sync.WaitGroup
	statuses := make(chan int, 100)
	for range 100 {
		wg.Go(func() {
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, httptest.NewRequest("GET", "/", nil))
			statuses <- rec.Code
		})
	}
	wg.Wait()
	close(statuses)
	allowed, rejected := 0, 0
	for status := range statuses {
		switch status {
		case http.StatusNoContent:
			allowed++
		case http.StatusTooManyRequests:
			rejected++
		default:
			t.Fatalf("unexpected status: %d", status)
		}
	}
	if allowed != 10 || rejected != 90 {
		t.Fatalf("allowed = %d, rejected = %d; want 10, 90", allowed, rejected)
	}
}
