package httpx

import "net/http"

// NoStoreHeader sets Cache-Control to no-store on the response.
func NoStoreHeader(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
}
