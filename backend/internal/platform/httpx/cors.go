package httpx

import "net/http"

// CORS allows credentialed requests from one origin while retaining CSRF protection.
func CORS(origin string) (Middleware, error) {
	protection := http.NewCrossOriginProtection()
	if err := protection.AddTrustedOrigin(origin); err != nil {
		return nil, err
	}

	return func(next http.Handler) http.Handler {
		protected := protection.Handler(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Vary", "Origin")
			if r.Header.Get("Origin") == origin {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
					w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, POST, PUT, DELETE")
					w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}
			protected.ServeHTTP(w, r)
		})
	}, nil
}
