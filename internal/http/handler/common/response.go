package common

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/dgdraganov/crispy-chat-service/internal/model"
)

type writer struct {
	logger *slog.Logger
}

func NewWriter(logger *slog.Logger) *writer {
	return &writer{
		logger: logger,
	}
}

// Write is used by http handlers in order to write the specific
// message and status code to the response writer object
func (w *writer) Write(ctx context.Context, rw http.ResponseWriter, message string, statusCode int) {
	requestID := ctx.Value(model.RequestID).(string)

	respMsg := model.ResponseMessage{
		Message: message,
	}
	resp, err := json.Marshal(respMsg)
	if err != nil {
		w.logger.Error(
			"json marshal error",
			"request_id", requestID,
			"error", err,
		)
		rw.WriteHeader(http.StatusInternalServerError)
		if _, err := rw.Write([]byte("Something went wrong on our end!")); err != nil {
			w.logger.Error(
				"write response error",
				"request_id", requestID,
				"error", err,
			)
		}
		return
	}
	rw.WriteHeader(statusCode)
	if _, err := rw.Write(resp); err != nil {
		rw.WriteHeader(http.StatusInternalServerError)
		w.logger.Error(
			"write response error",
			"request_id", requestID,
			"error", err,
		)
	}
}
