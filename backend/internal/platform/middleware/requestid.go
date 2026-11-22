package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
)

type reqIDContextKey struct{}

func WithReqID(ctx context.Context, reqID string) context.Context {
	return context.WithValue(ctx, reqIDContextKey{}, reqID)
}

func ReqIDFromContext(ctx context.Context) (string, bool) {
	id, ok := ctx.Value(reqIDContextKey{}).(string)
	return id, ok
}

func SetRequestID() Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			reqID := r.Header.Get("X-Request-ID")
			if reqID == "" || len(reqID) > 64 {
				reqID = generateID()
			}
			w.Header().Set("X-Request-ID", reqID)

			ctx := WithReqID(r.Context(), reqID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func generateID() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
