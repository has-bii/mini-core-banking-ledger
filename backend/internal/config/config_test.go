package config_test

import (
	"backend/internal/config"
	"os"
	"slices"
	"strings"
	"testing"
)

func TestLoad_CORSOrigin(t *testing.T) {
	setupEnv(t)
	for _, tc := range []struct {
		name, value string
		want        []string
	}{
		{"empty defaults", "", []string{"http://localhost:3000"}},
		{"single origin", "https://app.example.com", []string{"https://app.example.com"}},
		{"multiple origins", "https://app.example.com,https://admin.example.com", []string{"https://app.example.com", "https://admin.example.com"}},
		{"trim and skip blanks", " , https://app.example.com, ,\thttps://admin.example.com ,", []string{"https://app.example.com", "https://admin.example.com"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("CORS_ORIGIN", tc.value)
			cfg, err := config.Load()
			if err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(cfg.CORSOrigin, tc.want) {
				t.Fatalf("CORSOrigin = %q, want %q", cfg.CORSOrigin, tc.want)
			}
		})
	}
}

func TestLoad_UnsetCORSOrigin(t *testing.T) {
	setupEnv(t)
	if err := os.Unsetenv("CORS_ORIGIN"); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(cfg.CORSOrigin, []string{"http://localhost:3000"}) {
		t.Fatalf("CORSOrigin = %q, want localhost default", cfg.CORSOrigin)
	}
}

func TestLoad_InvalidCORSOrigin(t *testing.T) {
	setupEnv(t)
	for _, tc := range []struct {
		name, value string
	}{
		{"whitespace only", " \t "},
		{"commas only", ",,,"},
		{"commas and whitespace", " , \t, "},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("CORS_ORIGIN", tc.value)
			_, err := config.Load()
			if err == nil || !strings.Contains(err.Error(), "CORS_ORIGIN") {
				t.Fatalf("Load() error = %v, want error naming CORS_ORIGIN", err)
			}
		})
	}
}

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
		"CORS_ORIGIN":            "",
		"HTTP_READ_TIMEOUT":      "5s",
		"SHUTDOWN_TIMEOUT":       "10s",
		"RATE_LIMIT_CAPACITY":    "",
		"RATE_LIMIT_REFILL_RATE": "",
	} {
		t.Setenv(key, value)
	}
}
