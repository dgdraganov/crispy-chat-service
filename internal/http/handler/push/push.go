package push

import (
	"context"
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

	clientID := q.Get("client_id")
	if clientID == "" {
		writer.Write(
			r.Context(),
			w,
			"Missing client_id query parameter!",
			http.StatusBadRequest,
		)
		return
	}

	ctx := context.WithValue(r.Context(), model.ClientID, clientID)

	valid, err := handler.publisher.Verify(signature, clientID)
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
		"client_id", clientID,
	)

	msgChan := make(chan string)
	go handler.readingMessages(ctx, conn, msgChan)

	for {
		select {
		case <-r.Context().Done():
			handler.closingConnection(ctx, conn)
			return
		case msg, ok := <-msgChan:
			if !ok {
				return
			}
			err = handler.publisher.Publish(ctx, clientID, msg)
			if err != nil {
				handler.logger.Error(
					"publish message failed",
					"request_id", requestID,
					"error", err,
				)
				writer.Write(
					ctx,
					w,
					"Something went wrong on our end!",
					http.StatusInternalServerError,
				)
				continue
			}
			handler.logger.Info(
				"message published successfully",
				"request_id", requestID,
				"client_id", clientID,
			)
		}
	}
}

func (handler *pushHandler) closingConnection(ctx context.Context, conn *websocket.Conn) {
	requestID := ctx.Value(model.RequestID).(string)

	err := conn.WriteMessage(websocket.CloseMessage, []byte("closing connection"))
	if err != nil {
		handler.logger.Error(
			"conn writing message",
			"request_id", requestID,
			"error", err,
		)
	}
	conn.Close()
}

func (handler *pushHandler) readingMessages(ctx context.Context, conn *websocket.Conn, msgChan chan string) {
	requestID := ctx.Value(model.RequestID).(string)

Loop:
	for {
		msgType, b, err := conn.ReadMessage()
		if err != nil {
			handler.logger.Error(
				"read message error",
				"request_id", requestID,
				"error", err,
			)
			break
		}
		if msgType == websocket.CloseMessage {
			handler.logger.Info(
				"request for closing ws connection",
				"request_id", requestID,
			)
			break
		}

		select {
		case <-ctx.Done():
			break Loop
		case msgChan <- string(b):
			<-time.After(time.Millisecond * 100)
			continue
		}

	}
	close(msgChan)
	conn.Close()
}
