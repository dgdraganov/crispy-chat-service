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
	logs      *slog.Logger
}

// NewHandler is a cpnstructor function for the authHandler type
func NewHandler(method string, publisher Publisher, logger *slog.Logger) *pushHandler {
	return &pushHandler{
		publisher: publisher,
		logs:      logger,
		method:    method,
	}
}

// ServeHTTP implements the http.Handler interface for the authHandler type
func (handler *pushHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value(model.RequestID).(string)
	signature := r.Context().Value(model.Signature).(string)

	if r.Method != handler.method {
		handler.logs.Warn(
			"invalid request method for push endpoint",
			"request_method", r.Method,
			"expected_request_method", handler.method,
			"request_id", requestID,
		)
		err := common.WriteResponse(
			w,
			fmt.Sprintf("invalid request method %s, expected method is %s", r.Method, handler.method),
			http.StatusBadRequest,
		)
		if err != nil {
			handler.logs.Error(
				"write response failed",
				"request_id", requestID,
				"error", err,
			)
		}
		return
	}

	pushReq := model.PushRequest{}

	// decode json request body
	err := json.NewDecoder(r.Body).Decode(&pushReq)
	if err != nil {
		handler.logs.Error(
			"json decode",
			"request_id", requestID,
			"error", err,
		)
		err := common.WriteResponse(
			w,
			"invalid JSON request body",
			http.StatusBadRequest,
		)
		if err != nil {
			handler.logs.Error(
				"write response",
				"request_id", requestID,
				"error", err,
			)
		}
		return
	}

	valid, err := handler.publisher.Verify(signature, pushReq.ClientID)
	if errors.Is(err, core.ErrInvalidIDFormat) {
		err := common.WriteResponse(
			w,
			"Invalid client ID",
			http.StatusBadRequest,
		)
		if err != nil {
			handler.logs.Error(
				"write response",
				"request_id", requestID,
				"error", err,
			)
		}
		return
	}
	if err != nil {
		handler.logs.Error(
			"publisher verify",
			"request_id", requestID,
			"error", err,
		)
		err := common.WriteResponse(
			w,
			"Something went wrong on our end!",
			http.StatusInternalServerError,
		)
		if err != nil {
			handler.logs.Error(
				"write response",
				"request_id", requestID,
				"error", err,
			)
		}
	}
	if !valid {
		err := common.WriteResponse(
			w,
			"Invalid signature",
			http.StatusBadRequest,
		)
		if err != nil {
			handler.logs.Error(
				"write response",
				"request_id", requestID,
				"error", err,
			)
		}
		return
	}

	err = handler.publisher.Publish(r.Context(), pushReq.ClientID, pushReq.Message)
	if err != nil {
		handler.logs.Error(
			"publisher publish",
			"request_id", requestID,
			"error", err,
		)
		err := common.WriteResponse(
			w,
			"Something went wrong on our end!",
			http.StatusInternalServerError,
		)
		if err != nil {
			handler.logs.Error(
				"write response",
				"request_id", requestID,
				"error", err,
			)
		}
		return
	}

	err = common.WriteResponse(
		w,
		"Message published successfully!",
		http.StatusOK,
	)
	if err != nil {
		handler.logs.Error(
			"write response",
			"request_id", requestID,
			"error", err,
		)
	}
}
