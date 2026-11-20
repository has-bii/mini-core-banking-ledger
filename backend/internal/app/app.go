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
)

func New(cfg config.Config, pool *pgxpool.Pool, log *slog.Logger) (http.Handler, error) {
	cors, err := middleware.CORS(&middleware.CORSOptions{
		Origin:         cfg.CORSOrigin,
		AllowedHeaders: "Content-Type",
		AllowedMethods: "GET, HEAD, POST, PUT, DELETE",
		Credentials:    true,
	})
	if err != nil {
		return nil, fmt.Errorf("CORS_ORIGIN: %w", err)
	}

	tb := middleware.NewRateLimiter(cfg.RatelimitCapacity, cfg.RatelimitRefillRate)

	sessionSvc := session.NewService(session.NewStore(pool))
	authSvc := auth.NewService(user.NewStore(pool), sessionSvc)
	walletService := wallet.NewService(wallet.NewStore(pool))

	lookupUserID := func(ctx context.Context, token string) (string, error) {
		current, err := sessionSvc.FindByToken(ctx, token)
		if errors.Is(err, session.ErrNotFound) {
			return "", nil
		}
		return current.UserID, err
	}

	mux := http.NewServeMux()
	api := http.NewServeMux()

	auth.NewHandler(authSvc, log, cfg.Env != "dev" && cfg.Env != "test").Register(api)
	wallet.NewHandler(walletService, log).Register(api)

	mux.Handle("/api/", http.StripPrefix("/api", api))

	return middleware.Chain(
		mux,
		middleware.Recoverer(log),
		middleware.Logging(log),
		cors,
		middleware.GetUserFromSession(lookupUserID),
		tb.Middleware(),
	), nil
}
