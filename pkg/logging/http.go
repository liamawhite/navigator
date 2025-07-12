package logging

import (
	"log/slog"
	"net/http"
	"time"
)

// HTTPMiddleware returns an HTTP middleware that logs requests and responses
func HTTPMiddleware() func(http.Handler) http.Handler {

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID := GenerateRequestID()

			// Create request-scoped logger
			requestLogger := For(ComponentHTTP).With(
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"user_agent", r.UserAgent(),
				"remote_addr", r.RemoteAddr,
			)

			// Add request ID to response headers for client debugging
			w.Header().Set("X-Request-ID", requestID)

			// Create a response writer wrapper to capture status code
			wrapped := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}

			requestLogger.Debug("http request started")

			// Process the request
			next.ServeHTTP(wrapped, r)

			// Calculate duration
			duration := time.Since(start)

			// Log the result
			logLevel := slog.LevelInfo
			if wrapped.statusCode >= 400 {
				logLevel = slog.LevelWarn
			}
			if wrapped.statusCode >= 500 {
				logLevel = slog.LevelError
			}

			requestLogger.Log(r.Context(), logLevel, "http request completed",
				"status_code", wrapped.statusCode,
				"duration_ms", duration.Milliseconds(),
				"response_size", wrapped.bytesWritten,
			)
		})
	}
}

// responseWriter wraps http.ResponseWriter to capture status code and response size
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	bytesWritten int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	n, err := rw.ResponseWriter.Write(b)
	rw.bytesWritten += n
	return n, err
}
