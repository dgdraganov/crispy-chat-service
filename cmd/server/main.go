package main

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/dgdraganov/crispy-chat-service/internal/config"
	"github.com/dgdraganov/crispy-chat-service/internal/http/router"
	"github.com/dgdraganov/crispy-chat-service/internal/http/server"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	conf, err := config.NewServer()
	if err != nil {
		panic(fmt.Sprintf("load server config: %s", err))
	}

	router := router.New()

	//router.Register("/auth", ...)
	//router.Register("/auth", ...)
	//router.Register("/auth", ...)

	server := server.NewHTTP(conf.Port, router.ServeMux(), logger)
	// starts the server asynchronously
	server.Start(conf.Port)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-sig

	logger.Info("shut down signal received")
	server.Shutdown()
}
