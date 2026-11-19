package config_test

import (
	"backend/internal/config"
	"os"
	"strings"
	"testing"
)

func TestLoad_CORSOrigin(t *testing.T) {
	setupEnv(t)
	for _, tc := range []struct {
		name, value string
		want        string
	}{
		{"empty defaults", "", "http://localhost:3000"},
		{"single origin", "https://app.example.com", "https://app.example.com"},
		{"origin with port", "http://localhost:3001", "http://localhost:3001"},
		{"trim whitespace", " \thttps://app.example.com \t", "https://app.example.com"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("CORS_ORIGIN", tc.value)
			cfg, err := config.Load()
			if err != nil {
				t.Fatal(err)
			}
			if cfg.CORSOrigin != tc.want {
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
	if cfg.CORSOrigin != "http://localhost:3000" {
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
		{"multiple origins", "https://app.example.com,https://admin.example.com"},
		{"trailing comma", "https://app.example.com,"},
		{"comma in host", "https://app,admin.example.com"},
		{"wildcard", "*"},
		{"wildcard host", "https://*.example.com"},
		{"missing scheme", "app.example.com"},
		{"missing host", "https://"},
		{"path", "https://app.example.com/path"},
		{"query", "https://app.example.com?key=value"},
		{"fragment", "https://app.example.com#section"},
		{"malformed URL", "https://%"},
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

func TestLoad_TrimStringValues(t *testing.T) {
	setupEnv(t)
	for key, value := range map[string]string{
		"APP_ENV":      " \ttest \n",
		"HTTP_ADDR":    " :9090 ",
		"DATABASE_URL": " postgresql://unused/test ",
		"REDIS_URL":    " redis://unused:6379 ",
		"LOG_LEVEL":    " debug ",
	} {
		t.Setenv(key, value)
	}
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.Env != "test" || cfg.Addr != ":9090" || cfg.DatabaseURL != "postgresql://unused/test" || cfg.RedisURL != "redis://unused:6379" || cfg.LogLevel != "debug" {
		t.Fatalf("string values not trimmed: %+v", cfg)
	}
}

func TestLoad_WhitespaceOnlyRequiredStrings(t *testing.T) {
	for _, key := range []string{"DATABASE_URL", "REDIS_URL"} {
		t.Run(key, func(t *testing.T) {
			setupEnv(t)
			t.Setenv(key, " \t ")
			_, err := config.Load()
			if err == nil || !strings.Contains(err.Error(), key) {
				t.Fatalf("Load() error = %v, want error naming %s", err, key)
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
