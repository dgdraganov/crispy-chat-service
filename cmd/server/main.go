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
	l := middleware.NewLologMiddleware(logger)

	// Load private key
	privateKey, err := sign.LoadPrivateKey(os.Getenv("PRIVATE_KEY"))
	if err != nil {
		panic(fmt.Sprintf("load private key: %s", err))
	}

	// Instantiate core functionality
	dSigner := sign.NewECDSA(privateKey, sign.NewSHA256Hasher(), sign.NewBase64Encoder())
	redisStore := redis.New(conf.RedisAddress)
	chatCore := core.New(dSigner, redisStore, logger)

	// Handlers
	var authHandler http.Handler
	authHandler = auth.NewHandler("POST", chatCore, logger)
	authHandler = i.Id(l.Log(authHandler))

	var pushHandler http.Handler
	pushHandler = push.NewHandler("GET", chatCore, logger)
	pushHandler = i.Id(l.Log(a.Auth(pushHandler)))

	var listenHandler http.Handler
	listenHandler = listen.NewHandler("GET", chatCore, logger)
	listenHandler = i.Id(l.Log(a.Auth(listenHandler)))

	// Router
	mux := &http.ServeMux{}
	mux.Handle("/auth", authHandler)
	mux.Handle("/push", pushHandler)
	mux.Handle("/listen", listenHandler)

	server := server.NewHTTP(conf.Port, mux, logger)

	// Start the server asynchronously
	server.Start(conf.Port)

	// Graceful shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-sig

	logger.Info("shut down signal received")
	server.Shutdown()
}
