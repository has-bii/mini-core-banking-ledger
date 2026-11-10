package app

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"backend/internal/config"
	"backend/internal/modules/account"
	"backend/internal/modules/auth"
	"backend/internal/modules/session"
	"backend/internal/modules/user"
	"backend/internal/platform/transport/httpx/middleware"

	"github.com/jackc/pgx/v5/pgxpool"
)

func New(cfg config.Config, pool *pgxpool.Pool, log *slog.Logger) (http.Handler, error) {
	cors, err := middleware.CORS(cfg.CORSOrigin)
	if err != nil {
		return nil, fmt.Errorf("CORS_ORIGIN: %w", err)
	}

	tb := middleware.NewTokenBucket(cfg.RatelimitCapacity, cfg.RatelimitRefillRate)

	mux := http.NewServeMux()

	sessionSvc := session.NewService(session.NewStore(pool))
	lookupUserID := func(ctx context.Context, token string) (string, error) {
		current, err := sessionSvc.FindByToken(ctx, token)
		if errors.Is(err, session.ErrNotFound) {
			return "", nil
		}
		return current.UserID, err
	}
	authSvc := auth.NewService(pool, user.NewStore(pool), account.NewStore(pool), sessionSvc)

	api := http.NewServeMux()
	auth.NewHandler(authSvc, log, cfg.Env != "dev" && cfg.Env != "test").Register(api)
	mux.Handle("/api/", http.StripPrefix("/api", api))

	return middleware.Chain(mux, middleware.Recoverer(log), middleware.Logging(log), cors, middleware.GetUser(lookupUserID), tb.Middleware()), nil
}
