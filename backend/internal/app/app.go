package app

import (
	"log/slog"
	"net/http"

	"backend/internal/modules/auth"
	"backend/internal/modules/session"
	"backend/internal/platform/httpx"

	"github.com/jackc/pgx/v5/pgxpool"
)

func New(env string, pool *pgxpool.Pool, log *slog.Logger) http.Handler {

	mux := http.NewServeMux()

	sessionSvc := session.NewService(session.NewStore(pool))
	authSvc := auth.NewService(pool, sessionSvc)

	auth.NewHandler(authSvc, log, env != "dev" && env != "test").Register(mux)

	protected := http.NewCrossOriginProtection().Handler(mux)
	return httpx.Chain(protected, httpx.Recoverer(log), httpx.Logging(log))
}
