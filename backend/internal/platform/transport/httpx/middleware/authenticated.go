package middleware

import (
	"backend/internal/platform/transport/httpx"
	"context"
	"net/http"
)

type userIDContextKey struct{}

func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDContextKey{}).(string)
	return id, ok
}

// GetUser resolves a session cookie; an empty user ID means anonymous, errors return 500.
func GetUser(lookupUserID func(context.Context, string) (string, error)) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session")
			if err != nil || cookie.Value == "" {
				next.ServeHTTP(w, r)
				return
			}

			userID, err := lookupUserID(r.Context(), cookie.Value)
			if err != nil {
				httpx.WriteError(w, http.StatusInternalServerError, httpx.CodeInternalError, "internal error")
				return
			}
			if userID == "" {
				next.ServeHTTP(w, r)
				return
			}
			ctx := context.WithValue(r.Context(), userIDContextKey{}, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AuthGuard requires a verified user ID populated by GetUser.
func AuthGuard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id, ok := UserIDFromContext(r.Context()); !ok || id == "" {
			httpx.NoStoreHeader(w)
			httpx.WriteError(w, http.StatusUnauthorized, httpx.CodeUnauthenticated, "authentication required")
			return
		}
		next.ServeHTTP(w, r)
	})
}
