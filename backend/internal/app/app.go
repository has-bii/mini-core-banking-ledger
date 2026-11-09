package app

import (
	"log/slog"
	"net/http"

	"backend/internal/modules/auth"
	"backend/internal/modules/user"
	"backend/internal/platform/httpx"

	"github.com/jackc/pgx/v5/pgxpool"
)

func New(env string, pool *pgxpool.Pool, log *slog.Logger) http.Handler {

	mux := http.NewServeMux()

	userStore := user.NewStore(pool)
	authSvc := auth.NewService(userStore)

	auth.NewHandler(authSvc, log).Register(mux)

	protected := http.NewCrossOriginProtection().Handler(mux)
	return httpx.Chain(protected, httpx.Recoverer(log), httpx.Logging(log))
}
