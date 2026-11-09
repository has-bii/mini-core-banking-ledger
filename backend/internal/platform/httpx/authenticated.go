package httpx

import (
	"backend/internal/modules/session"
	"context"
	"errors"
	"net/http"
)

type sessionService interface {
	FindByToken(context.Context, string) (session.Session, error)
}

type userIDContextKey struct{}

func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDContextKey{}).(string)
	return id, ok
}

// GetUser resolves an optional session; missing or invalid sessions remain anonymous.
func GetUser(svc sessionService) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session")
			if err != nil || cookie.Value == "" {
				next.ServeHTTP(w, r)
				return
			}

			current, err := svc.FindByToken(r.Context(), cookie.Value)
			if errors.Is(err, session.ErrNotFound) {
				next.ServeHTTP(w, r)
				return
			}
			if err != nil {
				WriteError(w, http.StatusInternalServerError, CodeInternalError, "internal error")
				return
			}
			ctx := context.WithValue(r.Context(), userIDContextKey{}, current.UserID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// if UserID from context is null, return Unauthorized error
func AuthGuard() {

}
