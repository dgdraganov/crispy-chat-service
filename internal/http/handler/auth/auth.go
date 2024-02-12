package auth

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/dgdraganov/crispy-chat-service/internal/core"
	"github.com/dgdraganov/crispy-chat-service/internal/model"
	"github.com/dgdraganov/crispy-chat-service/pkg/common"
)

type authHandler struct {
	method string
	auth   Authenticator
	logger *slog.Logger
}

// NewHandler is a cpnstructor function for the authHandler type
func NewHandler(method string, authenticator Authenticator, logger *slog.Logger) *authHandler {
	return &authHandler{
		auth:   authenticator,
		logger: logger,
		method: method,
	}
}

// ServeHTTP implements the http.Handler interface for the authHandler type
func (handler *authHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value(model.RequestID).(string)

	writer := common.NewWriter(handler.logger)

	if r.Method != handler.method {
		handler.logger.Warn(
			"invalid request method for auth endpoint",
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

	authReq := model.AuthRequest{}

	// decode json request body
	err := json.NewDecoder(r.Body).Decode(&authReq)
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

	signature, err := handler.auth.Sign(authReq.ClientID)
	if errors.Is(err, core.ErrInvalidIDFormat) {
		writer.Write(
			r.Context(),
			w,
			"invalid format for clientID! Format example: eccbaa4f-5075-40d6-81c6-e3dbdd3bbf9a",
			http.StatusBadRequest,
		)

		return
	}
	if err != nil {
		handler.logger.Error(
			"auth sign error",
			"request_id", requestID,
			"error", err,
			"client_id", authReq.ClientID,
		)
		writer.Write(
			r.Context(),
			w,
			"Something went wrong on our end!",
			http.StatusInternalServerError,
		)
		return
	}

	writer.Write(
		r.Context(),
		w,
		signature,
		http.StatusOK,
	)
}
