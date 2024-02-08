package core

import (
	"errors"
	"fmt"
	"regexp"
)

var ErrWrongIDFormat error = errors.New("client ID not in the correct format")

type chatService struct {
	ds      DSigner
	idRegex string
}

func New(ds DSigner, idRegex string) *chatService {
	return &chatService{
		ds:      ds,
		idRegex: idRegex,
	}
}

func (chat *chatService) Authenticate(clientID string) (string, error) {
	match, err := regexp.MatchString(chat.idRegex, clientID)
	if err != nil {
		return "", fmt.Errorf("regex Match: %w", err)
	}
	if !match {
		return "", ErrWrongIDFormat
	}

	signature, err := chat.ds.Sign(clientID)
	if err != nil {
		return "", fmt.Errorf("digital sign: %w", err)
	}

	return signature, nil
}
