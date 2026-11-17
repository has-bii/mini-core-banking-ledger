package middleware

import (
	"backend/internal/platform/v2/httpx"
	"context"
	"net/http"
)

type userIDContextKey struct{}

func UserIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(userIDContextKey{}).(string)
	return id, ok
}

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
				httpx.SendError(w, httpx.NewErrorResponse(
					http.StatusText(http.StatusInternalServerError),
					http.StatusInternalServerError,
					"Unexpected error has occurred",
					r.URL.Path,
					nil,
				))
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

func AuthGuard(h func(http.ResponseWriter, *http.Request)) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if id, ok := UserIDFromContext(r.Context()); !ok || id == "" {
			httpx.NoStoreHeader(w)
			httpx.SendError(w, httpx.NewErrorResponse(
				http.StatusText(http.StatusUnauthorized),
				http.StatusUnauthorized,
				"Session is expired or invalid",
				r.URL.Path,
				nil,
			))
			return
		}
		h(w, r)
	})
}
