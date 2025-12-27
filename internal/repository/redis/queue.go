package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"red_collar/internal/domain"

	"github.com/redis/go-redis/v9"
)

const webhookQueueKey = "webhook:queue"

type Queue struct {
	client *redis.Client
}

func NewQueue(client *redis.Client) *Queue {
	return &Queue{
		client: client,
	}
}

func (q *Queue) Enqueue(ctx context.Context, check *domain.LocationCheck) error {
	data, err := json.Marshal(check)
	if err != nil {
		return fmt.Errorf("failed to marshal location check: %w", err)
	}

	if err := q.client.LPush(ctx, webhookQueueKey, data).Err(); err != nil {
		return fmt.Errorf("failed to enqueue location check: %w", err)
	}
	return nil
}

func (q *Queue) Dequeue(ctx context.Context) (*domain.LocationCheck, error) {
	result, err := q.client.BRPop(ctx, 0, webhookQueueKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to dequeue location check: %w", err)
	}

	if len(result) < 2 {
		return nil, fmt.Errorf("invalid queue result")
	}

	var check domain.LocationCheck
	if err := json.Unmarshal([]byte(result[1]), &check); err != nil {
		return nil, fmt.Errorf("failed to unmarshal location check: %w", err)
	}
	return &check, nil
}
