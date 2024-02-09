package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-redis/redis"
)

type redisStore struct {
	client *redis.Client
}

func New() *redisStore {
	client := redis.NewClient(&redis.Options{
		Addr:     "redis:6379",
		Password: "",
		DB:       0,
	})

	if err := client.Ping().Err(); err != nil {
		panic(fmt.Sprintf("ping redis client: %s", err))
	}

	return &redisStore{
		client: client,
	}
}

func (store *redisStore) PublishMessage(room string, sortKey float64, message any) (int64, error) {
	messageBytes, err := json.Marshal(message)
	if err != nil {
		return 0, fmt.Errorf("json marshal: %w", err)
	}

	setMember := redis.Z{Score: sortKey, Member: string(messageBytes)}

	result, err := store.client.ZAdd(room, setMember).Result()
	if err != nil {
		return 0, fmt.Errorf("redis ZAdd: %w", err)
	}

	return result, nil
}

func (store *redisStore) ReadMessages(ctx context.Context, room string) (<-chan string, <-chan error) {
	msgsChan := make(chan string)
	errorsChan := make(chan error)
	var offset int64
	var maxMessages int64 = 100

	go func() {
		for {
			// todo: add context.Done() for function return
			messages, err := store.client.ZRange(room, offset, offset+maxMessages).Result()
			if err != nil {
				errorsChan <- fmt.Errorf("redis ZRange: %w", err)
				<-time.After(500)
			}
			if len(messages) == 0 {
				<-time.After(500)
				continue
			}
			offset += int64(len(messages))
			for _, msg := range messages {
				msgsChan <- msg
			}
		}
	}()
	return msgsChan, errorsChan
}
