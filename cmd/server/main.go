package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/dgdraganov/crispy-chat-service/core"
	"github.com/dgdraganov/crispy-chat-service/internal/config"
	"github.com/dgdraganov/crispy-chat-service/internal/http/handler/auth"
	"github.com/dgdraganov/crispy-chat-service/internal/http/middleware"
	"github.com/dgdraganov/crispy-chat-service/internal/http/router"
	"github.com/dgdraganov/crispy-chat-service/internal/http/server"
	"github.com/dgdraganov/crispy-chat-service/pkg/sign"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	conf, err := config.NewServer()
	if err != nil {
		panic(fmt.Sprintf("load server config: %s", err))
	}

	// middleware initialization
	i := middleware.NewRequestIdMiddleware(logger)

	privateKey, err := sign.LoadPrivateKey(conf.PrivateKeyPath)
	if err != nil {
		logger.Error(
			"failed loading private key",
			"error", err,
		)
	}
	dSigner := sign.NewECDSA(privateKey, sign.NewSHA256Hasher(), sign.NewBase64Encoder())
	coreChat := core.New(dSigner, "^[a-z0-9]{8}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{12}$")

	var authHandler http.Handler
	authHandler = auth.NewHandler("POST", coreChat, logger)
	authHandler = i.Id(authHandler)

	router := router.New()
	router.Register("/auth", authHandler)
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
