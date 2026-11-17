package httpx

import "net/http"

func NoStoreHeader(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store")
}
