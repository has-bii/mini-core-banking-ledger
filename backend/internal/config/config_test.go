package config_test

import (
	"backend/internal/config"
	"strings"
	"testing"
)

func TestLoad_RateLimit(t *testing.T) {
	setupEnv(t)
	for _, tc := range []struct {
		name, capacity, refill   string
		wantCapacity, wantRefill float64
	}{
		{"defaults", "", "", 50, 10},
		{"custom", "75.5", "2.5", 75.5, 2.5},
		{"minimum capacity", "1", "0.001", 1, 0.001},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("RATE_LIMIT_CAPACITY", tc.capacity)
			t.Setenv("RATE_LIMIT_REFILL_RATE", tc.refill)
			cfg, err := config.Load()
			if err != nil {
				t.Fatal(err)
			}
			if cfg.RatelimitCapacity != tc.wantCapacity || cfg.RatelimitRefillRate != tc.wantRefill {
				t.Fatalf("rate limit = (%g, %g), want (%g, %g)", cfg.RatelimitCapacity, cfg.RatelimitRefillRate, tc.wantCapacity, tc.wantRefill)
			}
		})
	}
}

func TestLoad_InvalidRateLimit(t *testing.T) {
	setupEnv(t)
	for _, tc := range []struct {
		name, key, value string
	}{
		{"malformed", "RATE_LIMIT_CAPACITY", "invalid"},
		{"trailing text", "RATE_LIMIT_REFILL_RATE", "10oops"},
		{"overflow", "RATE_LIMIT_CAPACITY", "1e999"},
		{"zero capacity", "RATE_LIMIT_CAPACITY", "0"},
		{"zero refill", "RATE_LIMIT_REFILL_RATE", "0"},
		{"negative capacity", "RATE_LIMIT_CAPACITY", "-1"},
		{"negative refill", "RATE_LIMIT_REFILL_RATE", "-1"},
		{"NaN", "RATE_LIMIT_CAPACITY", "NaN"},
		{"positive infinity", "RATE_LIMIT_REFILL_RATE", "+Inf"},
		{"negative infinity", "RATE_LIMIT_REFILL_RATE", "-Inf"},
		{"capacity below one request", "RATE_LIMIT_CAPACITY", "0.5"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv(tc.key, tc.value)
			_, err := config.Load()
			if err == nil || !strings.Contains(err.Error(), tc.key) {
				t.Fatalf("Load() error = %v, want error naming %s", err, tc.key)
			}
		})
	}
}

func setupEnv(t *testing.T) {
	t.Helper()
	t.Chdir(t.TempDir())
	for key, value := range map[string]string{
		"APP_ENV":                "test",
		"DATABASE_URL":           "postgresql://unused/test",
		"REDIS_URL":              "redis://unused:6379",
		"LOG_LEVEL":              "info",
		"HTTP_READ_TIMEOUT":      "5s",
		"SHUTDOWN_TIMEOUT":       "10s",
		"RATE_LIMIT_CAPACITY":    "",
		"RATE_LIMIT_REFILL_RATE": "",
	} {
		t.Setenv(key, value)
	}
}
