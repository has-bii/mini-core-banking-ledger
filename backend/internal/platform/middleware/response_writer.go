package middleware

import "net/http"

type responseWriter struct {
	http.ResponseWriter
	committed bool
	status    int
}

// WriteHeader forwards statuses until the first final status, which it
// records. Informational statuses other than 101 do not commit the response.
func (w *responseWriter) WriteHeader(status int) {
	if w.committed {
		return
	}

	if status == http.StatusSwitchingProtocols || status >= 200 {
		w.committed = true
		w.status = status
	}

	w.ResponseWriter.WriteHeader(status)
}

// Write records an implicit 200 if no final status has been sent, then writes p.
func (w *responseWriter) Write(p []byte) (int, error) {
	if !w.committed {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(p)
}

// Unwrap returns the underlying writer for http.ResponseController.
func (w *responseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
