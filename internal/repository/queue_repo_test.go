package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"red_collar/internal/domain"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestQueueRepository_Enqueue(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	if testQueueRepo == nil {
		setupQueue()
	}

	ctx := context.Background()

	tests := []struct {
		name     string
		input    *domain.LocationCheck
		validate func(t *testing.T)
	}{
		{
			name: "success",
			input: &domain.LocationCheck{
				ID:     1,
				UserID: "colorvax",
			},
			validate: func(t *testing.T) {
				result, err := testRD.LPop(ctx, "webhook:queue").Result()
				require.NoError(t, err)

				var webhookTask WebhookTask
				err = json.Unmarshal([]byte(result), &webhookTask)
				require.NoError(t, err)

				require.NotNil(t, webhookTask)
				require.Equal(t, 1, webhookTask.LocationCheck.ID)
				require.Equal(t, "colorvax", webhookTask.LocationCheck.UserID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanupTestRD(t)

			err := testQueueRepo.Enqueue(ctx, tt.input)
			require.NoError(t, err)

			if tt.validate != nil {
				tt.validate(t)
			}
		})
	}
}

func TestQueueRepository_EnqueueWithDelay(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	if testQueueRepo == nil {
		setupQueue()
	}

	ctx := context.Background()

	tests := []struct {
		name     string
		task     *WebhookTask
		delay    time.Duration
		validate func(t *testing.T)
	}{
		{
			name: "success",
			task: &WebhookTask{
				LocationCheck: &domain.LocationCheck{
					ID:     1,
					UserID: "colorvax",
				},
			},
			delay: 1 * time.Second,
			validate: func(t *testing.T) {
				time.Sleep(1 * time.Second)
				tasks, err := testRD.ZRangeByScore(ctx, "webhook:delayed", &redis.ZRangeBy{
					Min:   "0",
					Max:   fmt.Sprintf("%d", time.Now().Unix()),
					Count: 100,
				}).Result()
				require.NoError(t, err)

				require.Len(t, tasks, 1)

				var webhook WebhookTask
				err = json.Unmarshal([]byte(tasks[0]), &webhook)
				require.NoError(t, err)

				require.Equal(t, 1, webhook.LocationCheck.ID)
				require.Equal(t, "colorvax", webhook.LocationCheck.UserID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanupTestRD(t)

			err := testQueueRepo.EnqueueWithDelay(ctx, tt.task, tt.delay)
			require.NoError(t, err)

			if tt.validate != nil {
				tt.validate(t)
			}
		})
	}
}

func TestQueueRepository_ProcessDelayedTasks(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	if testQueueRepo == nil {
		setupQueue()
	}

	ctx := context.Background()

	tests := []struct {
		name     string
		setup    func(t *testing.T)
		validate func(t *testing.T)
	}{
		{
			name: "success",
			setup: func(t *testing.T) {
				task := &WebhookTask{
					LocationCheck: &domain.LocationCheck{
						ID:     1,
						UserID: "colorvax",
					},
				}

				err := testQueueRepo.EnqueueWithDelay(ctx, task, 0)
				require.NoError(t, err)
			},
			validate: func(t *testing.T) {
				tasks, err := testRD.ZRangeByScore(ctx, "webhook:delayed", &redis.ZRangeBy{
					Min:   "0",
					Max:   fmt.Sprintf("%d", time.Now().Unix()),
					Count: 100,
				}).Result()
				require.NoError(t, err)
				require.Len(t, tasks, 0)

				res, err := testRD.LPop(ctx, "webhook:queue").Result()
				require.NoError(t, err)

				var webhook WebhookTask
				err = json.Unmarshal([]byte(res), &webhook)
				require.NoError(t, err)

				require.Equal(t, 1, webhook.LocationCheck.ID)
				require.Equal(t, "colorvax", webhook.LocationCheck.UserID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanupTestRD(t)

			if tt.setup != nil {
				tt.setup(t)
			}

			err := testQueueRepo.ProcessDelayedTasks(ctx)
			require.NoError(t, err)

			if tt.validate != nil {
				tt.validate(t)
			}
		})
	}
}

func TestQueueRepository_Dequeue(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	if testQueueRepo == nil {
		setupQueue()
	}

	ctx := context.Background()

	tests := []struct {
		name     string
		setup    func(t *testing.T)
		validate func(t *testing.T, res *WebhookTask)
	}{
		{
			name: "success",
			setup: func(t *testing.T) {
				err := testQueueRepo.Enqueue(ctx, &domain.LocationCheck{
					ID:     1,
					UserID: "colorvax",
				})
				require.NoError(t, err)
			},
			validate: func(t *testing.T, res *WebhookTask) {
				_, err := testRD.LPop(ctx, "webhook:queue").Result()
				require.Equal(t, redis.Nil, err)

				require.Equal(t, 1, res.LocationCheck.ID)
				require.Equal(t, "colorvax", res.LocationCheck.UserID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanupTestRD(t)

			if tt.setup != nil {
				tt.setup(t)
			}

			res, err := testQueueRepo.Dequeue(ctx)
			require.NoError(t, err)

			if tt.validate != nil {
				tt.validate(t, res)
			}
		})
	}
}
