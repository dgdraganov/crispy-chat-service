package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
)

type httpServer struct {
	logger *slog.Logger
	server *http.Server
}

// NewHTTP is a constructor function for the httpServer type
func NewHTTP(port string, serveMux *http.ServeMux, logger *slog.Logger) *httpServer {
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", port),
		Handler: serveMux,
	}

	return &httpServer{
		logger: logger,
		server: server,
	}
}

// Start runs an http server on the specified port
func (s *httpServer) Start(port string) {
	go func() {
		s.logger.Info(
			"Server starting...",
			"service_port", port,
		)
		if err := s.server.ListenAndServe(); err != nil {
			s.logger.Error(
				"server failed unexpectedly",
				"error", err,
			)
			os.Exit(1)
		}
	}()
}

// Shutdown stops the server gracefully
func (s *httpServer) Shutdown() {

	if err := s.server.Shutdown(context.Background()); err != nil {
		slog.Error(
			"server failed unexpectedly",
			"error", err,
		)
	}
}
