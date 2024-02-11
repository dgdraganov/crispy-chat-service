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
	"syscall"
	"time"

	"github.com/dgdraganov/crispy-chat-service/internal/model"
	"github.com/gorilla/websocket"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	host := os.Getenv("SERVICE_HOST")
	port := os.Getenv("SERVICE_PORT")
	clientID := os.Getenv("CLIENT_ID")

	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "9205"
	}
	if clientID == "" {
		randGen := rand.New(rand.NewSource(time.Now().UnixNano()))
		randomInt := randGen.Intn(89) + 10
		clientID = fmt.Sprintf("client_%d", randomInt)
	}

	authUrl := fmt.Sprintf("http://%s:%s/auth", host, port)
	listenUrl := fmt.Sprintf("ws://%s:%s/listen?client_id=%s", host, port, clientID)

	postBody, _ := json.Marshal(map[string]string{
		"client_id": clientID,
	})
	client := http.Client{}
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
	fmt.Println(authMsg.Message)
	headers := map[string][]string{
		"Signature": {authMsg.Message},
	}
	c, _, err := websocket.DefaultDialer.Dial(listenUrl, headers)
	if err != nil {
		log.Fatal("ws dial:", err)
	}

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

Loop:
	for {
		select {
		case <-sig:
			c.Close()
			logger.Info(
				"closing chat connection",
			)
			break Loop
		default:
			_, message, err := c.ReadMessage()
			if err != nil {
				logger.Error(
					"ws connection read:",
					"error", err,
				)
				continue
			}
			fmt.Printf("%s", string(message))
		}
	}

}
