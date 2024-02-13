package listen

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/dgdraganov/crispy-chat-service/internal/core"
	"github.com/dgdraganov/crispy-chat-service/internal/model"
	"github.com/dgdraganov/crispy-chat-service/pkg/common"
	"github.com/gorilla/websocket"
)

type listenHandler struct {
	method   string
	listener Listener
	logger   *slog.Logger
}

// NewHandler is a cpnstructor function for the listenHandler type
func NewHandler(method string, listener Listener, logger *slog.Logger) *listenHandler {
	return &listenHandler{
		listener: listener,
		logger:   logger,
		method:   method,
	}
}

// ServeHTTP implements the http.Handler interface for the listenHandler type
func (handler *listenHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	requestID := r.Context().Value(model.RequestID).(string)
	signature := r.Context().Value(model.Signature).(string)

	writer := common.NewWriter(handler.logger)

	if r.Method != handler.method {
		handler.logger.Warn(
			"invalid request method for /push endpoint",
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

	valid, err := handler.listener.Verify(signature, clientId)
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

	msgChan := handler.listener.ReadMessages(r.Context(), clientId)

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
	defer conn.Close()

	handler.logger.Info(
		"upgraded to websockets",
		"request_id", requestID,
		"client_id", clientId,
	)

	done := make(chan struct{})
	go func() {
		for {
			mt, _, err := conn.ReadMessage()
			if err != nil {
				handler.logger.Error(
					"conn read message",
					"error", err,
				)
				done <- struct{}{}
				break
			}
			if mt == websocket.CloseMessage {
				handler.logger.Info("sending close message")
				done <- struct{}{}
				break
			}
		}
	}()

Loop:
	for {
		select {
		case msg := <-msgChan:
			if err := conn.WriteMessage(websocket.TextMessage, []byte(msg)); err != nil {
				handler.logger.Error(
					"write message error",
					"request_id", requestID,
					"error", err,
				)
				break Loop
			}
		case <-done:
			break Loop
		case <-r.Context().Done():
			handler.logger.Info(
				"closing connection",
				"request_id", requestID,
			)
			if err := conn.WriteMessage(websocket.CloseMessage, []byte("closing connection")); err != nil {
				handler.logger.Error(
					"sending close message failed",
					"request_id", requestID,
					"error", err,
				)
			}
			break Loop
		default:
			<-time.After(time.Millisecond * 100)
		}

	}
}
