package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/dgdraganov/crispy-chat-service/internal/model"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	msgGen := MessageGenerator()

	host := os.Getenv("SERVICE_HOST")
	port := os.Getenv("SERVICE_PORT")
	name := os.Getenv("BOT_NAME")

	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "9205"
	}
	if name == "" {
		randGen := rand.New(rand.NewSource(time.Now().UnixNano()))
		randomInt := randGen.Intn(89) + 10
		name = fmt.Sprintf("bot_%d", randomInt)
	}

	client := http.Client{}
	authUrl := fmt.Sprintf("http://%s:%s/auth", host, port)
	pushUrl := fmt.Sprintf("http://%s:%s/push", host, port)

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

	for {
		postBody, _ := json.Marshal(map[string]string{
			"client_id": name,
			"message":   msgGen(),
		})
		responseBody := bytes.NewBuffer(postBody)
		req, err := http.NewRequest("POST", pushUrl, responseBody)
		if err != nil {
			logger.Error(
				"http NewRequest",
				"error", err,
			)
			continue
		}
		req.Header.Add("Signature", authMsg.Message)
		resp, err := client.Do(req)
		if err != nil {
			logger.Error(
				"client post",
				"error", err,
				"url", pushUrl,
			)
			continue
		}
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Error(
				"io read all",
				"error", err,
			)
			continue
		}

		logger.Info(
			"server response",
			"body", string(b),
		)
		resp.Body.Close()
		<-time.After(time.Second * 2)
	}
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
