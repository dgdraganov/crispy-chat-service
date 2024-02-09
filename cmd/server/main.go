package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/dgdraganov/crispy-chat-service/internal/config"
	"github.com/dgdraganov/crispy-chat-service/internal/core"
	"github.com/dgdraganov/crispy-chat-service/internal/http/handler/auth"
	"github.com/dgdraganov/crispy-chat-service/internal/http/handler/listen"
	"github.com/dgdraganov/crispy-chat-service/internal/http/handler/push"
	"github.com/dgdraganov/crispy-chat-service/internal/http/middleware"
	"github.com/dgdraganov/crispy-chat-service/internal/http/router"
	"github.com/dgdraganov/crispy-chat-service/internal/http/server"
	"github.com/dgdraganov/crispy-chat-service/pkg/redis"
	"github.com/dgdraganov/crispy-chat-service/pkg/sign"
)

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	conf, err := config.NewServer()
	if err != nil {
		panic(fmt.Sprintf("load server config: %s", err))
	}

	// Middleware
	i := middleware.NewRequestIdMiddleware(logger)
	a := middleware.NewSignatureMiddleware(logger)

	privateKey, err := sign.LoadPrivateKey(conf.PrivateKeyPath)
	if err != nil {
		panic("cannot load private key")
	}

	dSigner := sign.NewECDSA(privateKey, sign.NewSHA256Hasher(), sign.NewBase64Encoder())
	redisStore := redis.New()
	chatCore := core.New(dSigner, redisStore, logger)

	// Handlers
	var authHandler http.Handler
	authHandler = auth.NewHandler("POST", chatCore, logger)
	authHandler = i.Id(authHandler)

	var pushHandler http.Handler
	pushHandler = push.NewHandler("POST", chatCore, logger)
	pushHandler = i.Id(a.Auth(pushHandler))

	var listenHandler http.Handler
	listenHandler = listen.NewHandler("GET", chatCore, logger)
	listenHandler = i.Id(a.Auth(listenHandler))

	// Router
	router := router.New()
	router.Register("/auth", authHandler)
	router.Register("/push", pushHandler)
	router.Register("/listen", listenHandler)

	server := server.NewHTTP(conf.Port, router.ServeMux(), logger)
	// starts the server asynchronously
	server.Start(conf.Port)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-sig

	logger.Info("shut down signal received")
	server.Shutdown()
}
