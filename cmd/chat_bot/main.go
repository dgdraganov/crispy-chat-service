package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/dgdraganov/crispy-chat-service/internal/model"
	"github.com/gorilla/websocket"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	genMessage := MessageGenerator()

	host := os.Getenv("SERVICE_HOST")
	port := os.Getenv("SERVICE_PORT")
	name := os.Getenv("BOT_NAME")

	if host == "" {
		host = "chat-server"
	}
	if port == "" {
		port = "9205"
	}
	if name == "" {
		randGen := rand.New(rand.NewSource(time.Now().UnixNano()))
		randomInt := randGen.Intn(89) + 10
		name = fmt.Sprintf("bot_%d", randomInt)
	}

	authUrl := fmt.Sprintf("http://%s:%s/auth", host, port)
	pushUrl := fmt.Sprintf("ws://%s:%s/push?client_id=%s", host, port, name)

	// Authenticate with the server``
	client := http.Client{}
	postBody, _ := json.Marshal(map[string]string{
		"client_id": name,
	})
	responseBody := bytes.NewBuffer(postBody)
	resp, err := client.Post(authUrl, "application/json", responseBody)
	if err != nil {
		logger.Error(
			"client post",
			"error", err,
			"url", authUrl,
		)
		return
	}
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		logger.Error(
			"io read all",
			"error", err,
		)
		return
	}
	authMsg := model.ResponseMessage{}
	err = json.Unmarshal(b, &authMsg)
	if err != nil {
		logger.Error(
			"json unmarshal",
			"error", err,
		)
		return
	}

	// websocket connection
	headers := map[string][]string{
		"Signature": {authMsg.Message},
	}
	conn, _, err := websocket.DefaultDialer.Dial(pushUrl, headers)
	if err != nil {
		log.Fatal("ws dial:", err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

Loop:
	for {
		select {
		case <-sig:
			if err := conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "woops")); err != nil {
				logger.Error(
					"sending close conn message failed",
					"error", err,
				)
			}
			break Loop
		default:
			newMessage := genMessage()
			if err := conn.WriteMessage(websocket.TextMessage, []byte(newMessage)); err != nil {
				logger.Error(
					"conn write message",
					"error", err,
				)
				return
			}
			logger.Info(
				"message sent successfully",
				"client_id", name,
			)
			<-time.After(time.Second * 4)
		}
	}
	//conn.Close()
	logger.Info(
		"terminating chat connection",
	)
}

func MessageGenerator() func() string {
	randGen := rand.New(rand.NewSource(time.Now().UnixNano()))

	var loremIpsum string = "lorem ipsum dolor sit amet consectetur adipiscing elit sed do eiusmod tempor incididunt ut labore et dolore magna aliqua ut enim ad minim veniam quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur excepteur sint occaecat cupidatat non proident sunt in culpa qui officia deserunt mollit anim id est laborum"
	words := strings.Split(loremIpsum, " ")

	return func() string {
		i := randGen.Intn(len(words) - 5)
		message := strings.Join(words[i:i+5], " ")
		return message
	}
}
