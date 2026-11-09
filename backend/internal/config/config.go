package config

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Env, Addr, DatabaseURL, LogLevel string
	ReadTimeout, ShutdownTimeout     time.Duration
}

func Load() (Config, error) {
	if err := godotenv.Load(); err != nil && !errors.Is(err, os.ErrNotExist) {
		return Config{}, fmt.Errorf("load .env: %w", err)
	}
	var errs []error

	cfg := Config{
		Env:             getString("APP_ENV", "dev"),
		Addr:            getString("HTTP_ADDR", ":8080"),
		DatabaseURL:     getString("DATABASE_URL", ""),
		LogLevel:        getString("LOG_LEVEL", "info"),
		ReadTimeout:     getDuration("HTTP_READ_TIMEOUT", 5*time.Second, &errs),
		ShutdownTimeout: getDuration("SHUTDOWN_TIMEOUT", 10*time.Second, &errs),
	}

	if cfg.DatabaseURL == "" {
		errs = append(errs, errors.New("DATABASE_URL is required"))
	}

	switch cfg.Env {
	case "dev", "staging", "prod":
	default:
		errs = append(errs, fmt.Errorf("APP_ENV must be dev, staging, or prod; got %q", cfg.Env))
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
