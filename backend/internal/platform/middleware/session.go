package middleware

import (
	"backend/internal/platform/httpx"
	"context"
	"net/http"
)

type userIDContextKey struct{}

// WithUserID carries a verified user ID across middleware boundaries.
func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, userIDContextKey{}, userID)
}

func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDContextKey{}).(string)
	return id, ok
}

// GetUserFromSession logs lookup failures, so lookup errors must not contain the raw token.
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
