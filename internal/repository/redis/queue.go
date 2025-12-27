package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"red_collar/internal/domain"
	"time"

	"github.com/redis/go-redis/v9"
)

const (
	webhookQueueKey   = "webhook:queue"
	webhookDelayedKey = "webhook:delayed"
	webhookDLQKey     = "webhook:dlq"
)

type WebhookTask struct {
	LocationCheck *domain.LocationCheck `json:"location_check"`
	Attempt       int                   `json:"attempt"`
	FirstAttempt  time.Time             `json:"first_attempt"`
	LastError     string                `json:"last_error,omitempty"`
}

type Queue struct {
	client *redis.Client
}

func NewQueue(client *redis.Client) *Queue {
	return &Queue{
		client: client,
	}
}

// Добавление таска в обычную очередь
func (q *Queue) Enqueue(ctx context.Context, check *domain.LocationCheck) error {
	task := &WebhookTask{
		LocationCheck: check,
		Attempt:       0,
		FirstAttempt:  time.Now(),
	}

	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal location check: %w", err)
	}

	if err := q.client.LPush(ctx, webhookQueueKey, data).Err(); err != nil {
		return fmt.Errorf("failed to enqueue location check: %w", err)
	}
	return nil
}

// Добавление таска в отложенную очередь
func (q *Queue) EnqueueWithDelay(ctx context.Context, task *WebhookTask, delay time.Duration) error {
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook task: %w", err)
	}

	executeAt := time.Now().Add(delay).Unix()
	if err := q.client.ZAdd(ctx, webhookDelayedKey, redis.Z{
		Score:  float64(executeAt),
		Member: data,
	}).Err(); err != nil {
		return fmt.Errorf("failed to enqueue delayed webhook task: %w", err)
	}
	return nil
}

// Добавление таска в dlq
func (q *Queue) EnqueueDLQ(ctx context.Context, task *WebhookTask) error {
	data, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal webhook task: %w", err)
	}

	if err := q.client.LPush(ctx, webhookDLQKey, data).Err(); err != nil {
		return fmt.Errorf("failed to enqueue to DLQ: %w", err)
	}
	return nil
}

// Логика обработки отложенный задач
func (q *Queue) ProcessDelayedTasks(ctx context.Context) error {
	now := time.Now().Unix()

	tasks, err := q.client.ZRangeByScore(ctx, webhookDelayedKey, &redis.ZRangeBy{
		Min:   "0",
		Max:   fmt.Sprintf("%d", now),
		Count: 100,
	}).Result()

	if err != nil {
		return fmt.Errorf("failed to get delayed tasks: %w", err)
	}

	for _, taskData := range tasks {
		if err := q.client.ZRem(ctx, webhookDelayedKey, taskData).Err(); err != nil {
			continue
		}

		if err := q.client.LPush(ctx, webhookQueueKey, taskData).Err(); err != nil {
			q.client.ZAdd(ctx, webhookDelayedKey, redis.Z{
				Score:  float64(now + 5),
				Member: taskData,
			})
			continue
		}
	}
	return nil
}

// Получение таска из очереди
func (q *Queue) Dequeue(ctx context.Context) (*WebhookTask, error) {
	result, err := q.client.BRPop(ctx, 0, webhookQueueKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to dequeue webhook task: %w", err)
	}

	if len(result) < 2 {
		return nil, fmt.Errorf("invalid queue result")
	}

	var task WebhookTask
	if err := json.Unmarshal([]byte(result[1]), &task); err != nil {
		return nil, fmt.Errorf("failed to unmarshal webhook task: %w", err)
	}
	return &task, nil
}
