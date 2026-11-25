package middleware

import (
	"backend/internal/platform/httpx"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// CORSOptions configures the allowed origin and CORS response headers.
type CORSOptions struct {
	// Origin is one exact trusted origin, including scheme and optional port,
	// for example "https://app.example.com" or "http://localhost:3000".
	// Comma-separated lists and wildcards such as "*" are not supported.
	Origin string
	// AllowedHeaders lists comma-separated request headers browsers may send,
	// for example "Content-Type, Authorization".
	AllowedHeaders string
	// AllowedMethods lists comma-separated methods browsers may use,
	// for example "GET, POST, OPTIONS". It does not restrict server routing.
	AllowedMethods string
	// Credentials permits credentialed cross-origin requests from trusted
	// origins when true, for example requests carrying cookies.
	Credentials bool
	// MaxAge sets the preflight cache duration, for example 10 * time.Minute.
	// Positive values are sent as whole seconds; non-positive values omit the header.
	MaxAge time.Duration
}

// CORS returns middleware that sets CORS headers, answers OPTIONS with 204,
// and blocks unsafe cross-origin requests from untrusted origins.
// Pass non-nil opts and leave them unchanged. An invalid origin returns an error.
func CORS(opts *CORSOptions) (Middleware, error) {
	// Cross-origin protection: block unsafe cross-origin requests,
	// except from the origin we explicitly allow.
	cop := http.NewCrossOriginProtection()
	cop.SetDenyHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		httpx.SendError(w, r, httpx.NewHTTPError(http.StatusForbidden, "Cross-origin request denied."))
	}))
	if strings.ContainsAny(opts.Origin, ",*") {
		return nil, fmt.Errorf("invalid CORS origin %q: must be a single exact origin without wildcards", opts.Origin)
	}
	if err := cop.AddTrustedOrigin(opts.Origin); err != nil {
		return nil, fmt.Errorf("invalid CORS origin %q: %w", opts.Origin, err)
	}

	var maxAge string
	if opts.MaxAge > 0 {
		maxAge = strconv.Itoa(int(opts.MaxAge.Seconds()))
	}

	return func(next http.Handler) http.Handler {
		protected := cop.Handler(next)

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// The response depends on Origin, so caches must know that.
			w.Header().Add("Vary", "Origin")

			if origin == opts.Origin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				if opts.Credentials {
					w.Header().Set("Access-Control-Allow-Credentials", "true")
				}
			}

			w.Header().Set("Access-Control-Allow-Methods", opts.AllowedMethods)
			w.Header().Set("Access-Control-Allow-Headers", opts.AllowedHeaders)

			// Preflight is answered before the protection check.
			if r.Method == http.MethodOptions {
				// Let the browser cache this preflight result.
				if maxAge != "" {
					w.Header().Set("Access-Control-Max-Age", maxAge)
				}
				w.WriteHeader(http.StatusNoContent)
				return
			}

			protected.ServeHTTP(w, r)
		})
	}, nil
}
