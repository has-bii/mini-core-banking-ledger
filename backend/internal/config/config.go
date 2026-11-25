package config

import (
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

var (
	ErrDatabaseURLRequired = errors.New("DATABASE_URL is required")
	ErrRedisURLRequired    = errors.New("REDIS_URL is required")
)

type Config struct {
	Env, Addr, DatabaseURL, RedisURL, LogLevel    string
	CORSOrigin                                    string
	ReadTimeout, ShutdownTimeout, SessionLifetime time.Duration
	RatelimitCapacity, RatelimitRefillRate        float64
}

// Load reads and validates configuration from the environment, loading .env
// if present without overriding existing variables. It returns a zero Config
// on error and reports validation failures together.
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
		SessionLifetime:     getDuration("SESSION_LIFETIME", 30*time.Minute, &errs),
		RatelimitCapacity:   getFloat("RATE_LIMIT_CAPACITY", 50, &errs),
		RatelimitRefillRate: getFloat("RATE_LIMIT_REFILL_RATE", 10, &errs),
	}

	if strings.ContainsAny(cfg.CORSOrigin, ",*") {
		errs = append(errs, errors.New("CORS_ORIGIN must be a single exact origin without wildcards"))
	} else if err := http.NewCrossOriginProtection().AddTrustedOrigin(cfg.CORSOrigin); err != nil {
		errs = append(errs, fmt.Errorf("CORS_ORIGIN: %w", err))
	}

	if cfg.SessionLifetime%time.Second != 0 {
		errs = append(errs, errors.New("SESSION_LIFETIME must be a whole number of seconds"))
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

// getString trims nonempty environment values and uses the fallback only if
// the raw value is unset or empty.
func getString(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return strings.TrimSpace(value)
	}
	return fallback
}

// getDuration reads a positive duration. It returns the fallback for an empty
// or invalid value and appends an error only for invalid values.
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

// getFloat reads a finite positive number. It returns the fallback for an empty
// or invalid value and appends an error only for invalid values.
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
