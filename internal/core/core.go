package core

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/dgdraganov/crispy-chat-service/internal/model"
)

var ErrInvalidIDFormat error = errors.New("client ID not in the correct format")
var ErrInvalidSignature error = errors.New("client signature not valid")

type chatService struct {
	dSigner  DigitalSigner
	dbClient DBClient
	idRegex  string
	logger   *slog.Logger
	//chatRoom string
}

func New(ds DigitalSigner, dbClient DBClient, logger *slog.Logger) *chatService {
	//room := uuid.New().String()
	return &chatService{
		dSigner:  ds,
		dbClient: dbClient,
		logger:   logger,
		//chatRoom: room,
	}
}

func (chat *chatService) Sign(clientID string) (string, error) {
	match, err := regexp.MatchString(chat.idRegex, clientID)
	if err != nil {
		return "", fmt.Errorf("regex Match: %w", err)
	}
	if !match {
		return "", ErrInvalidIDFormat
	}

	signature, err := chat.dSigner.Sign(clientID)
	if err != nil {
		return "", fmt.Errorf("digital sign: %w", err)
	}

	return signature, nil
}

func (chat *chatService) Publish(ctx context.Context, clientID, message string) error {
	if message == "" {
		return errors.New("message is empty")
	}
	timestamp := time.Now().Unix()
	msgObj := model.ChatMessage{
		Timestamp: timestamp,
		Message:   message,
		ClientID:  clientID,
	}
	_, err := chat.dbClient.PublishMessage("common_room", float64(timestamp), msgObj)
	if err != nil {
		return fmt.Errorf("db client publish message: %w", err)
	}

	return nil
}

func (chat *chatService) ReadMessages(ctx context.Context, clientID string) <-chan string {
	requestID := ctx.Value(model.RequestID).(string)

	messagesChan := make(chan string)

	go func() {
		messages, errors := chat.dbClient.ReadMessages(ctx, "common_room")
		for {
			select {
			case msg := <-messages:
				msgObj := model.ChatMessage{}
				err := json.Unmarshal([]byte(msg), &msgObj)
				if err != nil {
					chat.logger.Error(
						"message unmarshal failed - discarding message",
						"error", err,
						"message_struct", msg,
						"request_id", requestID,
					)
					continue
				}
				msgTime := time.Unix(msgObj.Timestamp, 0).Format("15:04")
				chatMessage := fmt.Sprintf("[%s]\t%s\t%s\n", msgTime, clientID, msgObj.Message)
				messagesChan <- chatMessage
			case err := <-errors:
				chat.logger.Error(
					"redis read messages",
					"error", err,
					"request_id", requestID,
				)
				continue
			}
		}
	}()
	return messagesChan
}

func (chat *chatService) Verify(signature, clientID string) (bool, error) {
	if len(clientID) == 0 || strings.Contains(clientID, " ") {
		return false, ErrInvalidIDFormat
	}

	valid, err := chat.dSigner.Verify(signature, clientID)
	if err != nil {
		return false, fmt.Errorf("signature verify: %w", err)
	}
	return valid, nil
}
