package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

var logger *zap.Logger

// InitLogger initializes the middleware logger
func InitLogger(l *zap.Logger) {
	logger = l
}

func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Wrap response writer to capture status code
		wrapped := &responseWrapper{ResponseWriter: w, status: http.StatusOK}

		next.ServeHTTP(wrapped, r)

		// Log using zap
		if logger != nil {
			logger.Info("HTTP Request",
				zap.String("method", r.Method),
				zap.String("path", r.URL.Path),
				zap.Int("status", wrapped.status),
				zap.Duration("duration", time.Since(start)),
				zap.String("remote_addr", r.RemoteAddr),
				zap.String("user_agent", r.UserAgent()),
			)
		}
	})
}

type responseWrapper struct {
	http.ResponseWriter
	status int
}

func (w *responseWrapper) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}