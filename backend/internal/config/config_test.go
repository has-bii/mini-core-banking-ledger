package config

import (
	"errors"
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
