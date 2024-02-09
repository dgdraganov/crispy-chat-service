package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/dgdraganov/crispy-chat-service/internal/model"
)

type signatureMiddleware struct {
	logs *slog.Logger
}

func NewSignatureMiddleware(logger *slog.Logger) *signatureMiddleware {
	return &signatureMiddleware{
		logs: logger,
	}
}

// Auth implements the middleware logic to attach the signature to the request context
func (r *signatureMiddleware) Auth(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		signature := r.Header.Get("Signature")
		ctx := context.WithValue(r.Context(), model.Signature, signature)
		r = r.WithContext(ctx)

		handler.ServeHTTP(w, r)
	})
}
