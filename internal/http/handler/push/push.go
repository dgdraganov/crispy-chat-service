package push

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/dgdraganov/crispy-chat-service/internal/core"
	"github.com/dgdraganov/crispy-chat-service/internal/http/handler/common"
	"github.com/dgdraganov/crispy-chat-service/internal/model"
)

type pushHandler struct {
	method    string
	publisher Publisher
	logger    *slog.Logger
}

// NewHandler is a cpnstructor function for the authHandler type
func NewHandler(method string, publisher Publisher, logger *slog.Logger) *pushHandler {
	return &pushHandler{
		publisher: publisher,
		logger:    logger,
		method:    method,
	}
}

// ServeHTTP implements the http.Handler interface for the authHandler type
func (handler *pushHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value(model.RequestID).(string)
	signature := r.Context().Value(model.Signature).(string)

	writer := common.NewWriter(handler.logger)

	if r.Method != handler.method {
		handler.logger.Warn(
			"invalid request method for push endpoint",
			"request_method", r.Method,
			"expected_request_method", handler.method,
			"request_id", requestID,
		)
		writer.Write(
			r.Context(),
			w,
			fmt.Sprintf("invalid request method %s, expected method is %s", r.Method, handler.method),
			http.StatusBadRequest,
		)
		return
	}

	pushReq := model.PushRequest{}

	// decode json request body
	err := json.NewDecoder(r.Body).Decode(&pushReq)
	if err != nil {
		handler.logger.Error(
			"json decode",
			"request_id", requestID,
			"error", err,
		)
		writer.Write(
			r.Context(),
			w,
			"invalid JSON request body",
			http.StatusBadRequest,
		)

		return
	}

	valid, err := handler.publisher.Verify(signature, pushReq.ClientID)
	if errors.Is(err, core.ErrInvalidIDFormat) {
		writer.Write(
			r.Context(),
			w,
			"Invalid client ID",
			http.StatusBadRequest,
		)
		return
	}
	if err != nil {
		handler.logger.Error(
			"publisher verify",
			"request_id", requestID,
			"error", err,
		)
		writer.Write(
			r.Context(),
			w,
			"Something went wrong on our end!",
			http.StatusInternalServerError,
		)
		return
	}
	if !valid {
		writer.Write(
			r.Context(),
			w,
			"Invalid signature",
			http.StatusBadRequest,
		)
		return
	}

	err = handler.publisher.Publish(r.Context(), pushReq.ClientID, pushReq.Message)
	if err != nil {
		handler.logger.Error(
			"publisher publish",
			"request_id", requestID,
			"error", err,
		)
		writer.Write(
			r.Context(),
			w,
			"Something went wrong on our end!",
			http.StatusInternalServerError,
		)
		return
	}

	handler.logger.Info(
		"message published successfully",
		"request_id", requestID,
		"client_id", pushReq.ClientID,
	)

	writer.Write(
		r.Context(),
		w,
		"Message published successfully!",
		http.StatusOK,
	)
}
