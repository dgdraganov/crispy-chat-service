package middleware

import (
	"log/slog"
	"net/http"

	"github.com/dgdraganov/crispy-chat-service/internal/model"
)

type logMiddleware struct {
	logger *slog.Logger
}

func NewLologMiddleware(logger *slog.Logger) *logMiddleware {
	return &logMiddleware{
		logger: logger,
	}
}

// Log implements the middleware logic to log incoming http requests
func (l *logMiddleware) Log(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Context().Value(model.RequestID).(string)
		l.logger.Info(
			"incoming request",
			"method", r.Method,
			"request_id", requestID,
			"url", r.URL,
		)

		handler.ServeHTTP(w, r)
	})
}
