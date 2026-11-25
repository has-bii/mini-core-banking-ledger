package middleware

import (
	"backend/internal/platform/httpx"
	"context"
	"net/http"
)

type userIDContextKey struct{}

// WithUserID returns a context carrying the given verified user ID.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDContextKey{}, userID)
}

// UserIDFromContext returns the stored user ID and whether one was set.
// The boolean reports presence, not whether the ID is nonempty.
func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDContextKey{}).(string)
	return id, ok
}

// GetUserFromSession returns middleware that adds a user ID to the request
// context using the session cookie. It passes through without setting an ID
// if the cookie is missing or empty, or if the lookup returns an empty ID.
// Lookup errors produce a 500 and are logged, so they must not include the raw token.
func GetUserFromSession(lookupUserID func(context.Context, string) (string, error)) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session")
			if err != nil || cookie.Value == "" {
				next.ServeHTTP(w, r)
				return
			}

			userID, err := lookupUserID(r.Context(), cookie.Value)
			if err != nil {
				httpx.SendError(w, r, httpx.WrapHTTPError(http.StatusInternalServerError, "Unexpected error has occurred", err))
				return
			}
			if userID == "" {
				next.ServeHTTP(w, r)
				return
			}
			ctx := WithUserID(r.Context(), userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// AuthGuard wraps h and responds with 401 Unauthorized and Cache-Control:
// no-store if the request context has no nonempty user ID.
func AuthGuard(h func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id, ok := UserIDFromContext(r.Context()); !ok || id == "" {
			httpx.NoStoreHeader(w)
			httpx.SendError(w, r, httpx.NewHTTPError(http.StatusUnauthorized, "Session is expired or invalid"))
			return
		}
		h(w, r)
	})
}
