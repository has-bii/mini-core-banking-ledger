package app

import (
	"fmt"
	"log/slog"
	"net/http"

	"backend/internal/modules/account"
	"backend/internal/modules/auth"
	"backend/internal/modules/session"
	"backend/internal/modules/user"
	"backend/internal/platform/httpx"

	"github.com/jackc/pgx/v5/pgxpool"
)

func New(env, corsOrigin string, pool *pgxpool.Pool, log *slog.Logger) (http.Handler, error) {
	cors, err := httpx.CORS(corsOrigin)
	if err != nil {
		return nil, fmt.Errorf("CORS_ORIGIN: %w", err)
	}

	mux := http.NewServeMux()

	sessionSvc := session.NewService(session.NewStore(pool))
	authSvc := auth.NewService(pool, user.NewStore(pool), account.NewStore(pool), sessionSvc)

	api := http.NewServeMux()
	auth.NewHandler(authSvc, log, env != "dev" && env != "test").Register(api)
	mux.Handle("/api/", http.StripPrefix("/api", api))

	return httpx.Chain(mux, httpx.Recoverer(log), httpx.Logging(log), cors), nil
}
