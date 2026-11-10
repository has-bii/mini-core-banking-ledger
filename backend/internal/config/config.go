package config

import (
	"errors"
	"fmt"
	"math"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

var (
	ErrDatabaseURLRequired = errors.New("DATABASE_URL is required")
	ErrRedisURLRequired    = errors.New("REDIS_URL is required")
)

type Config struct {
	Env, Addr, DatabaseURL, RedisURL, LogLevel, CORSOrigin string
	ReadTimeout, ShutdownTimeout                           time.Duration
	RatelimitCapacity, RatelimitRefillRate                 float64
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}
	var errs []error

	cfg := Config{
		Env:                 getString("APP_ENV", "dev"),
		Addr:                getString("HTTP_ADDR", ":8080"),
		DatabaseURL:         getString("DATABASE_URL", ""),
		RedisURL:            getString("REDIS_URL", ""),
		LogLevel:            getString("LOG_LEVEL", "info"),
		CORSOrigin:          getString("CORS_ORIGIN", "http://localhost:3000"),
		ReadTimeout:         getDuration("HTTP_READ_TIMEOUT", 5*time.Second, &errs),
		ShutdownTimeout:     getDuration("SHUTDOWN_TIMEOUT", 10*time.Second, &errs),
		RatelimitCapacity:   getFloat("RATE_LIMIT_CAPACITY", 50, &errs),
		RatelimitRefillRate: getFloat("RATE_LIMIT_REFILL_RATE", 10, &errs),
	}

	if cfg.RatelimitCapacity < 1 {
		errs = append(errs, errors.New("RATE_LIMIT_CAPACITY must be at least 1"))
	}

	if cfg.DatabaseURL == "" {
		errs = append(errs, ErrDatabaseURLRequired)
	}
	if cfg.RedisURL == "" {
		errs = append(errs, ErrRedisURLRequired)
	}

	switch cfg.Env {
	case "dev", "test", "staging", "prod":
	default:
		errs = append(errs, fmt.Errorf("APP_ENV must be dev, test, staging, or prod; got %q", cfg.Env))
	}

	switch cfg.LogLevel {
	case "debug", "info", "warn", "error":
	default:
		errs = append(errs, fmt.Errorf("LOG_LEVEL must be debug, info, warn, or error; got %q", cfg.LogLevel))
	}

	if err := errors.Join(errs...); err != nil {
		return Config{}, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

func getString(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getDuration(key string, fallback time.Duration, errs *[]error) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	parsedValue, err := time.ParseDuration(value)
	if err != nil {
		*errs = append(*errs, fmt.Errorf("%s: %w", key, err))
		return fallback
	}
	if parsedValue <= 0 {
		*errs = append(*errs, fmt.Errorf("%s must be positive", key))
		return fallback
	}
	return parsedValue
}

func getFloat(key string, fallback float64, errs *[]error) float64 {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	parsed, err := strconv.ParseFloat(v, 64)
	if err != nil {
		*errs = append(*errs, fmt.Errorf("%s: %w", key, err))
		return fallback
	}
	if parsed <= 0 || math.IsNaN(parsed) || math.IsInf(parsed, 0) {
		*errs = append(*errs, fmt.Errorf("%s must be finite and positive", key))
		return fallback
	}
	return parsed
}
