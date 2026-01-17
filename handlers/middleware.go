package handlers

import (
	"net/http"
	"time"
)

func LoggingMiddleware(next http.HandlerFunc, logger Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next(wrapped, r)

		duration := time.Since(start)
		durationMs := float64(duration.Nanoseconds()) / 1e6

		logger.Log(r.Method, r.RequestURI, r.RemoteAddr, wrapped.statusCode, durationMs)
	}
}

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}
