package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/go-redis/redis"
)

type redisStore struct {
	client *redis.Client
}

// New is a constructor function for the redisStore type
func New(redisAddres string) *redisStore {
	client := redis.NewClient(&redis.Options{
		Addr:     redisAddres,
		Password: "",
		DB:       0,
	})
	for {
		if err := client.Ping().Err(); err != nil {
			slog.Error(
				"error connection to redis",
				"error", err,
				"redis_address", redisAddres,
			)
			<-time.After(time.Second * 1)
			continue
		}
		break
	}

	return &redisStore{
		client: client,
	}
}

// PublishMessage saves a message in the current redis instance
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

// ReadMessages reads messages from the current redis instance
func (store *redisStore) ReadMessages(ctx context.Context, room string) (<-chan string, <-chan error) {
	msgsChan := make(chan string)
	errorsChan := make(chan error)
	var offset int64
	var maxMessages int64 = 100

	go func() {
		for {
			<-time.After(time.Millisecond * 200)
			select {
			case <-ctx.Done():
				close(msgsChan)
				close(errorsChan)
				return
			default:
				messages, err := store.client.ZRange(room, offset, offset+maxMessages).Result()
				if err != nil {
					errorsChan <- fmt.Errorf("redis ZRange: %w", err)
					continue
				}
				if len(messages) == 0 {
					continue
				}
				for _, msg := range messages {
					msgsChan <- msg
				}
				offset += int64(len(messages))
			}
		}
	}()
	return msgsChan, errorsChan
}
