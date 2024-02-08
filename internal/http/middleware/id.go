package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/dgdraganov/crispy-chat-service/internal/model"
	"github.com/google/uuid"
)

type requestIdMiddleware struct {
	logs *slog.Logger
}

func NewRequestIdMiddleware(logger *slog.Logger) *requestIdMiddleware {
	return &requestIdMiddleware{
		logs: logger,
	}
}

// Id implements the middleware logic to attach an unique request id (uuid) to the request context
func (r *requestIdMiddleware) Id(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := uuid.New()
		ctx := context.WithValue(r.Context(), model.RequestID, requestID.String())
		r = r.WithContext(ctx)

		handler.ServeHTTP(w, r)
	})
}
