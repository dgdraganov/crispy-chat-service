package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/dgdraganov/crispy-chat-service/core"
	"github.com/dgdraganov/crispy-chat-service/internal/http/handler/common"
	"github.com/dgdraganov/crispy-chat-service/internal/model"
)

type authHandler struct {
	method string
	auth   Authenticator
	logs   *slog.Logger
}

// NewHandler is a cpnstructor function for the authHandler type
func NewHandler(method string, authenticator Authenticator, logger *slog.Logger) *authHandler {
	return &authHandler{
		auth:   authenticator,
		logs:   logger,
		method: method,
	}
}

// ServeHTTP implements the http.Handler interface for the authHandler type
func (handler *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value(model.RequestID).(string)

	if r.Method != handler.method {
		handler.logs.Warn(
			"invalid request method for auth endpoint",
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
				"write response",
				"request_id", requestID,
				"error", err,
			)
		}
		return
	}

	authReq := model.AuthRequest{}

	// decode json request body
	err := json.NewDecoder(r.Body).Decode(&authReq)
	if err != nil {
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

	signature, err := handler.auth.Authenticate(requestID)
	if errors.Is(err, core.ErrWrongIDFormat) {
		err := common.WriteResponse(
			w,
			"invalid format for clientID! Format example: eccbaa4f-5075-40d6-81c6-e3dbdd3bbf9a",
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
		signature,
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
