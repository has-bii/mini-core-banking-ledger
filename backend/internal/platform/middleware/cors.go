package middleware

import (
	"backend/internal/platform/httpx"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"time"
)

// CORSOptions configures allowed origins and CORS response headers.
type CORSOptions struct {
	// Origin lists exact trusted origins, including scheme and optional port,
	// for example []string{"https://app.example.com", "http://localhost:3000"}.
	// Wildcards such as "*" are not supported.
	Origin []string
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

// CORS creates middleware that sets CORS response headers and uses
// http.CrossOriginProtection to reject unsafe cross-origin requests from
// untrusted origins. It answers all OPTIONS requests with 204 No Content
// before applying that protection, but only trusted origins receive an
// Access-Control-Allow-Origin header.
//
// opts must be non-nil. CORS returns an error if a trusted origin is invalid.
// Treat opts and its Origin slice as immutable after calling CORS.
//
// Example:
//
//	cors, err := CORS(&CORSOptions{
//		Origin:         []string{"https://app.example.com"},
//		AllowedHeaders: "Content-Type, Authorization",
//		AllowedMethods: "GET, POST, OPTIONS",
//		Credentials:    true,
//		MaxAge:         10 * time.Minute,
//	})
//	if err != nil {
//		return err
//	}
//	mux := http.NewServeMux()
//	handler := cors(mux)
//	return http.ListenAndServe(":8080", handler)
func CORS(opts *CORSOptions) (Middleware, error) {
	// Cross-origin protection: block unsafe cross-origin requests,
	// except from the origins we explicitly allow.
	cop := http.NewCrossOriginProtection()
	cop.SetDenyHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// A write failure cannot be replaced with another response.
		_ = httpx.SendError(w, httpx.NewErrorResponse(
			http.StatusText(http.StatusForbidden),
			http.StatusForbidden,
			"Cross-origin request denied.",
			r.URL.Path,
			nil,
		))
	}))
	for _, o := range opts.Origin {
		if err := cop.AddTrustedOrigin(o); err != nil {
			return nil, fmt.Errorf("invalid CORS origin %q: %w", o, err)
		}
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

			if slices.Contains(opts.Origin, origin) {
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
