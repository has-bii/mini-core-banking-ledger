package middleware

import "net/http"

type responseWriter struct {
	http.ResponseWriter
	committed bool
	status    int
}

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

func (w *responseWriter) Write(p []byte) (int, error) {
	if !w.committed {
		w.WriteHeader(http.StatusOK)
	}
	return w.ResponseWriter.Write(p)
}

func (w *responseWriter) Unwrap() http.ResponseWriter {
	return w.ResponseWriter
}
