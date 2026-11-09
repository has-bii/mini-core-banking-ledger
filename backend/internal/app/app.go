package app

import (
	"fmt"
	"log/slog"
	"net/http"

	"backend/internal/config"
	"backend/internal/modules/account"
	"backend/internal/modules/auth"
	"backend/internal/modules/session"
	"backend/internal/modules/user"
	"backend/internal/platform/httpx"
	"backend/internal/platform/ratelimit"

	"github.com/jackc/pgx/v5/pgxpool"
)

func New(cfg config.Config, pool *pgxpool.Pool, log *slog.Logger) (http.Handler, error) {
	cors, err := httpx.CORS(cfg.CORSOrigin)
	if err != nil {
		return nil, fmt.Errorf("CORS_ORIGIN: %w", err)
	}

	tb := ratelimit.New(20, 10)

	mux := http.NewServeMux()

	sessionSvc := session.NewService(session.NewStore(pool))
	authSvc := auth.NewService(pool, user.NewStore(pool), account.NewStore(pool), sessionSvc)

	api := http.NewServeMux()
	auth.NewHandler(authSvc, log, cfg.Env != "dev" && cfg.Env != "test").Register(api)
	mux.Handle("/api/", http.StripPrefix("/api", api))

	return httpx.Chain(mux, httpx.Recoverer(log), httpx.Logging(log), cors, httpx.GetUser(sessionSvc), tb.Middleware()), nil
}
