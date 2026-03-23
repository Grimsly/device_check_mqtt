package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

// Wrapper around http.ResponseWriter to capture the status code for logging as http.ResponseWriter doesn't expose the status code
type logResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

// Replace http.ResponseWriter's WriteHeader so that status code can be saved in our wrapper
func (rw *logResponseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func RequestLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		wrapped := &logResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		slog.Info("Request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", wrapped.statusCode,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	})
}
