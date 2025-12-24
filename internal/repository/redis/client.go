package redis

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient(ctx context.Context, cfg string) (*RedisClient, error) {
	client := redis.NewClient(&redis.Options{
		Addr: "addr",
	})

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis server: %w", err)
	}

	return &RedisClient{client: client}, nil
}
