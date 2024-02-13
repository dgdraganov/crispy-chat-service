package config

import (
	"errors"
	"fmt"
	"os"
)

var errMissingEnvVariable error = errors.New("environment variable not found")

type ServerConfig struct {
	Port         string
	RedisAddress string
	PrivateKey   string
}

const (
	serverPortEnvString   = "SERVER_PORT"
	redisArrdessEnvString = "REDIS_ADDRESS"
	privateKeyEnvString   = "PRIVATE_KEY"
)

// NewServerConfig is a constructor function for the ServerConfig type
func NewServer() (ServerConfig, error) {
	port, ok := os.LookupEnv(serverPortEnvString)
	if !ok {
		return ServerConfig{}, fmt.Errorf("%w: %s", errMissingEnvVariable, serverPortEnvString)
	}

	privKey, ok := os.LookupEnv(privateKeyEnvString)
	if !ok {
		return ServerConfig{}, fmt.Errorf("%w: %s", errMissingEnvVariable, privateKeyEnvString)
	}
	redisAddr, ok := os.LookupEnv(redisArrdessEnvString)
	if !ok {
		return ServerConfig{}, fmt.Errorf("%w: %s", errMissingEnvVariable, redisArrdessEnvString)
	}

	return ServerConfig{
		Port:         port,
		PrivateKey:   privKey,
		RedisAddress: redisAddr,
	}, nil
}
