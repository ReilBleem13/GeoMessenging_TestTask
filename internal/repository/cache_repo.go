package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	timeExpiration = 15 * time.Minute
)

type CacheRepository struct {
	db *redis.Client
}

func NewCacheRepository(db *redis.Client) *CacheRepository {
	return &CacheRepository{
		db: db,
	}
}

func (c *CacheRepository) Save(ctx context.Context, data []byte, key string) error {
	err := c.db.Set(ctx, key, data, timeExpiration).Err()
	if err != nil {
		return fmt.Errorf("failed to set data in cache: %w", err)
	}
	return nil
}

func (c *CacheRepository) Get(ctx context.Context, key string) ([]byte, error) {
	data, err := c.db.GetEx(ctx, key, timeExpiration).Bytes()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get data from cache: %w", err)
	}
	return data, nil
}

func (c *CacheRepository) Delete(ctx context.Context, key string) (bool, error) {
	deleted, err := c.db.Del(ctx, key).Result()
	if err != nil {
		return false, fmt.Errorf("failed to delete data from cache: %w", err)
	}
	return deleted == 1, nil
}
