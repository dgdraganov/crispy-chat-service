package config

import (
	"errors"
	"fmt"
	"os"
)

var errMissingEnvVariable error = errors.New("environment variable not found")

type ServerConfig struct {
	Port           string
	PrivateKeyPath string
	RedisAddress   string
}

const (
	serverPortEnvString             = "SERVER_PORT"
	privateKeyPathEnvString         = "PRIVATE_KEY_PATH"
	redisArrdessEnvString           = "REDIS_ADDRESS"
	privateKeyredisArrdessEnvString = "PRIVATE_KEY"
)

// NewServerConfig is a constructor function for the ServerConfig type
func NewServer() (ServerConfig, error) {
	port, ok := os.LookupEnv(serverPortEnvString)
	if !ok {
		return ServerConfig{}, fmt.Errorf("%w: %s", errMissingEnvVariable, serverPortEnvString)
	}

	pkPath, ok := os.LookupEnv(privateKeyPathEnvString)
	if !ok {
		return ServerConfig{}, fmt.Errorf("%w: %s", errMissingEnvVariable, privateKeyPathEnvString)
	}
	redisAddr, ok := os.LookupEnv(redisArrdessEnvString)
	if !ok {
		return ServerConfig{}, fmt.Errorf("%w: %s", errMissingEnvVariable, redisArrdessEnvString)
	}

	return ServerConfig{
		Port:           port,
		PrivateKeyPath: pkPath,
		RedisAddress:   redisAddr,
	}, nil
}
