package middleware

import (
	"net/http"
	"slices"
)

type Middleware func(http.Handler) http.Handler

// Chain wraps h so the first middleware runs first on each request.
// With no middleware, it returns h unchanged.
func Chain(h http.Handler, mws ...Middleware) http.Handler {
	for _, mw := range slices.Backward(mws) {
		h = mw(h)
	}
	return h
}
