package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"backend/internal/config"
	"backend/internal/modules/auth"
	"backend/internal/modules/session"
	"backend/internal/modules/user"
	"backend/internal/modules/wallet"
	"backend/internal/platform/middleware"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
)

func New(cfg config.Config, pool *pgxpool.Pool, redisClient *redis.Client, log *slog.Logger) (http.Handler, error) {
	cors, err := middleware.CORS(&middleware.CORSOptions{
		Origin:         cfg.CORSOrigin,
		AllowedHeaders: "Content-Type",
		AllowedMethods: "GET, HEAD, POST, PUT, DELETE",
		Credentials:    true,
	})
	if err != nil {
		return nil, fmt.Errorf("CORS_ORIGIN: %w", err)
	}

	rateLimiter := middleware.NewRateLimiter(cfg.RatelimitCapacity, cfg.RatelimitRefillRate)

	sessionSvc := session.NewService(session.NewStore(redisClient), cfg.SessionLifetime)
	authSvc := auth.NewService(user.NewStore(pool), sessionSvc)
	walletService := wallet.NewService(wallet.NewStore(pool))

	lookupUserID := func(ctx context.Context, token string) (string, error) {
		current, err := sessionSvc.GetByToken(ctx, token)
		if errors.Is(err, session.ErrNotFound) {
			return "", nil
		}
		return current.UserID, err
	}

	mux := http.NewServeMux()
	api := http.NewServeMux()

	auth.NewHandler(authSvc, cfg.Env != "dev" && cfg.Env != "test", cfg.SessionLifetime).Register(api)
	wallet.NewHandler(walletService).Register(api)

	mux.Handle("/api/", http.StripPrefix("/api", api))

	return middleware.Chain(
		mux,
		middleware.Logging(log),
		middleware.Recoverer(),
		cors,
		middleware.GetUserFromSession(lookupUserID),
		rateLimiter.Middleware(),
	), nil
}
