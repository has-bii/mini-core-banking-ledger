package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"backend/internal/app"
	"backend/internal/config"
	"backend/internal/platform/logger"
	"backend/internal/platform/postgres"
	"backend/internal/platform/redis"
)

// startupError keeps potentially sensitive causes out of logs without discarding them.
type startupError struct {
	message string
	cause   error
}

func (e startupError) Error() string { return e.message }
func (e startupError) Unwrap() error { return e.cause }

func main() {
	cfg, err := config.Load()
	if err != nil {
		slog.Error("fatal", "err", "failed to load configuration")
		os.Exit(1)
	}

	log := logger.New(cfg.Env, cfg.LogLevel)
	if err := run(cfg, log); err != nil {
		log.Error("fatal", "err", err)
		os.Exit(1)
	}
}

func run(cfg config.Config, log *slog.Logger) error {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Info("connecting to PostgreSQL")
	pool, err := postgres.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return startupError{message: "failed to connect to PostgreSQL", cause: err}
	}
	defer pool.Close()
	log.Info("PostgreSQL connected")

	log.Info("connecting to Redis")
	redisClient, err := redis.Open(ctx, cfg.RedisURL)
	if err != nil {
		return startupError{message: "failed to connect to Redis", cause: err}
	}
	defer func() {
		if err := redisClient.Close(); err != nil {
			log.Warn("close Redis client", "err", err)
		}
	}()
	log.Info("Redis connected")

	handler, err := app.New(cfg, pool, redisClient, log)
	if err != nil {
		return startupError{message: "failed to initialize HTTP handler", cause: err}
	}
	server := &http.Server{
		Addr:              cfg.Addr,
		Handler:           handler,
		ReadHeaderTimeout: cfg.ReadTimeout,
		ReadTimeout:       cfg.ReadTimeout,
		WriteTimeout:      cfg.ReadTimeout,
		IdleTimeout:       cfg.ShutdownTimeout,
	}

	defer func() {
		if err := server.Close(); err != nil {
			log.Warn("close HTTP server", "err", err)
		}
	}()

	errs := make(chan error, 1)
	log.Info("starting HTTP server", "addr", cfg.Addr)
	go func() {
		errs <- server.ListenAndServe()
	}()

	select {
	case err := <-errs:
		if !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve HTTP: %w", err)
		}
		return nil
	case <-ctx.Done():
		log.Info("starting graceful HTTP shutdown")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			return fmt.Errorf("shut down HTTP server: %w", err)
		}
		log.Info("HTTP shutdown complete")
		return nil
	}
}
