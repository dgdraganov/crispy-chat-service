package push

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/dgdraganov/crispy-chat-service/internal/core"
	"github.com/dgdraganov/crispy-chat-service/internal/model"
	"github.com/dgdraganov/crispy-chat-service/pkg/common"
	"github.com/gorilla/websocket"
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

	q, err := url.ParseQuery(r.URL.RawQuery)
	if err != nil {
		handler.logger.Error(
			"url parse query",
			"request_id", requestID,
			"error", err,
		)
		writer.Write(
			r.Context(),
			w,
			"Invalid URL!",
			http.StatusBadRequest,
		)
		return
	}

	clientId := q.Get("client_id")
	if clientId == "" {
		writer.Write(
			r.Context(),
			w,
			"Missing client_id query parameter!",
			http.StatusBadRequest,
		)
		return
	}

	valid, err := handler.publisher.Verify(signature, clientId)
	if errors.Is(err, core.ErrInvalidIDFormat) {
		writer.Write(
			r.Context(),
			w,
			"Invalid client ID!",
			http.StatusBadRequest,
		)
		return
	}
	if err != nil {
		handler.logger.Error(
			"listener verify",
			"request_id", requestID,
			"error", err,
		)
		writer.Write(
			r.Context(),
			w,
			"Something went wrong on our end!",
			http.StatusInternalServerError,
		)
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

	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}

	upgrader.CheckOrigin = func(r *http.Request) bool { return true }

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		handler.logger.Error(
			"upgrade to websocet",
			"request_id", requestID,
			"error", err,
		)
		return
	}

	handler.logger.Info(
		"upgraded to websockets",
		"request_id", requestID,
		"client_id", clientId,
	)

	for {
		msgType, b, err := conn.ReadMessage()
		if err != nil {
			handler.logger.Error(
				"write message error",
				"request_id", requestID,
				"error", err,
			)
			conn.Close()
			return
		}

		if msgType == websocket.CloseMessage {
			conn.Close()
			handler.logger.Info(
				"gracefully terminating ws connection",
				"request_id", requestID,
			)
			break
		}

		err = handler.publisher.Publish(r.Context(), clientId, string(b))
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
			"client_id", clientId,
		)
	}

}
