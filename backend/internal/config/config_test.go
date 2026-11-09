package config

import (
	"errors"
	"strings"
	"testing"
)

func TestLoadRedisURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgresql://localhost/test")
	t.Setenv("APP_ENV", "test")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("HTTP_READ_TIMEOUT", "5s")
	t.Setenv("SHUTDOWN_TIMEOUT", "10s")

	t.Setenv("REDIS_URL", "redis://localhost:6379/1")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RedisURL != "redis://localhost:6379/1" {
		t.Fatalf("RedisURL=%q, want redis://localhost:6379/1", cfg.RedisURL)
	}

	t.Setenv("REDIS_URL", "")
	if _, err := Load(); !errors.Is(err, ErrRedisURLRequired) {
		t.Fatalf("expected missing REDIS_URL error, got %v", err)
	}
}

func TestLoadRateLimit(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgresql://localhost/test")
	t.Setenv("REDIS_URL", "redis://localhost:6379/0")
	t.Setenv("APP_ENV", "test")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("HTTP_READ_TIMEOUT", "5s")
	t.Setenv("SHUTDOWN_TIMEOUT", "10s")
	t.Setenv("RATE_LIMIT_RPS", "")
	t.Setenv("RATE_LIMIT_BURST", "")
	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if cfg.RateLimitRPS != 100 || cfg.RateLimitBurst != 200 {
		t.Fatalf("defaults = %d/%d, want 100/200", cfg.RateLimitRPS, cfg.RateLimitBurst)
	}
	t.Setenv("RATE_LIMIT_RPS", "25")
	t.Setenv("RATE_LIMIT_BURST", "50")
	cfg, err = Load()
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
				if _, err := Load(); err == nil || !strings.Contains(err.Error(), key) {
					t.Fatalf("expected %s validation error, got %v", key, err)
				}
			})
		}
	}
}

func TestLoadCORSOrigin(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgresql://localhost/test")
	t.Setenv("REDIS_URL", "redis://localhost:6379/0")
	t.Setenv("APP_ENV", "test")
	t.Setenv("LOG_LEVEL", "info")
	t.Setenv("HTTP_READ_TIMEOUT", "5s")
	t.Setenv("SHUTDOWN_TIMEOUT", "10s")
	for _, tt := range []struct{ value, want string }{
		{"", "http://localhost:3000"},
		{"https://app.example.com", "https://app.example.com"},
	} {
		t.Setenv("CORS_ORIGIN", tt.value)
		cfg, err := Load()
		if err != nil {
			t.Fatal(err)
		}
		if cfg.CORSOrigin != tt.want {
			t.Errorf("CORSOrigin=%q, want %q", cfg.CORSOrigin, tt.want)
		}
	}
}
