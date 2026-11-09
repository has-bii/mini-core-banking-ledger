package config_test

import (
	"errors"
	"strings"
	"testing"

	"backend/internal/config"
)

func configEnv(t *testing.T) {
	t.Helper()
	t.Chdir(t.TempDir())
	for key, value := range map[string]string{
		"DATABASE_URL":      "postgresql://localhost/test",
		"REDIS_URL":         "redis://localhost:6379/0",
		"APP_ENV":           "test",
		"LOG_LEVEL":         "info",
		"HTTP_READ_TIMEOUT": "5s",
		"SHUTDOWN_TIMEOUT":  "10s",
		"RATE_LIMIT_RPS":    "",
		"RATE_LIMIT_BURST":  "",
		"CORS_ORIGIN":       "",
	} {
		t.Setenv(key, value)
	}
}

func TestLoadRedisURL(t *testing.T) {
	configEnv(t)

	t.Setenv("REDIS_URL", "redis://localhost:6379/1")
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RedisURL != "redis://localhost:6379/1" {
		t.Fatalf("RedisURL=%q, want redis://localhost:6379/1", cfg.RedisURL)
	}

	t.Setenv("REDIS_URL", "")
	if _, err := config.Load(); !errors.Is(err, config.ErrRedisURLRequired) {
		t.Fatalf("expected missing REDIS_URL error, got %v", err)
	}
}

func TestLoadRateLimit(t *testing.T) {
	configEnv(t)
	cfg, err := config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RateLimitRPS != 100 || cfg.RateLimitBurst != 200 {
		t.Fatalf("defaults = %d/%d, want 100/200", cfg.RateLimitRPS, cfg.RateLimitBurst)
	}
	t.Setenv("RATE_LIMIT_RPS", "25")
	t.Setenv("RATE_LIMIT_BURST", "50")
	cfg, err = config.Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RateLimitRPS != 25 || cfg.RateLimitBurst != 50 {
		t.Fatalf("overrides = %d/%d, want 25/50", cfg.RateLimitRPS, cfg.RateLimitBurst)
	}
	for _, key := range []string{"RATE_LIMIT_RPS", "RATE_LIMIT_BURST"} {
		for _, value := range []string{"0", "-1", "abc", "1.5", "NaN", "Inf", "999999999999999999999999999999"} {
			t.Run(key+"/"+value, func(t *testing.T) {
				t.Setenv(key, value)
				if _, err := config.Load(); err == nil || !strings.Contains(err.Error(), key) {
					t.Fatalf("expected %s validation error, got %v", key, err)
				}
			})
		}
	}
}

func TestLoadCORSOrigin(t *testing.T) {
	configEnv(t)
	for _, tt := range []struct{ value, want string }{
		{"", "http://localhost:3000"},
		{"https://app.example.com", "https://app.example.com"},
	} {
		t.Setenv("CORS_ORIGIN", tt.value)
		cfg, err := config.Load()
		if err != nil {
			t.Fatal(err)
		}
		if cfg.CORSOrigin != tt.want {
			t.Errorf("CORSOrigin=%q, want %q", cfg.CORSOrigin, tt.want)
		}
	}
}
