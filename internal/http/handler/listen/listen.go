package listen

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/dgdraganov/crispy-chat-service/internal/core"
	"github.com/dgdraganov/crispy-chat-service/internal/http/handler/common"
	"github.com/dgdraganov/crispy-chat-service/internal/model"
	"github.com/gorilla/websocket"
)

type listenHandler struct {
	method   string
	listener Listener
	logs     *slog.Logger
}

// NewHandler is a cpnstructor function for the listenHandler type
func NewHandler(method string, listener Listener, logger *slog.Logger) *listenHandler {
	return &listenHandler{
		listener: listener,
		logs:     logger,
		method:   method,
	}
}

// ServeHTTP implements the http.Handler interface for the listenHandler type
func (handler *listenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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

	listenReq := model.ListenRequest{}
	// decode json request body
	err := json.NewDecoder(r.Body).Decode(&listenReq)
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

	valid, err := handler.listener.Verify(signature, listenReq.ClientID)
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
			"listener verify",
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

	msgChan := handler.listener.ReadMessages(r.Context(), listenReq.ClientID)

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		handler.logs.Error(
			"websocket upgrade",
			"request_id", requestID,
			"error", err,
		)
		return
	}

	handler.logs.Info(
		"upgraded to websockets",
		"request_id", requestID,
	)

	for {
		select {
		case msg := <-msgChan:
			if err := ws.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
				handler.logs.Error(
					"write message",
					"request_id", requestID,
					"error", err,
				)
			}
		case <-r.Context().Done():
			handler.logs.Info(
				"closing connection",
				"request_id", requestID,
			)
			return
		}

	}
}
