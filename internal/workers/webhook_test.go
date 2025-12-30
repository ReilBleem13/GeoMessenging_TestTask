package worker

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"red_collar/internal/domain"
	"red_collar/internal/repository"
	redisRepo "red_collar/internal/repository/redis"
	"red_collar/internal/service"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	redisC "github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/theartofdevel/logging"
)

var (
	wbOnce sync.Once

	rdContainer  *redisC.RedisContainer
	rdHost       string
	rdTest       *redisRepo.RedisClient
	rdTestClient *redis.Client

	queueTestRepo *repository.Queue
)

func createTestLogger() service.LoggerInterfaces {
	logger := logging.NewLogger(
		logging.WithLevel("warn"),
		logging.WithIsJSON(false),
	)
	return logger
}

func createTestServer(t *testing.T, handler http.HandlerFunc) *httptest.Server {
	server := httptest.NewServer(handler)
	t.Cleanup(func() {
		server.Close()
	})
	return server
}

func setup() error {
	var err error

	wbOnce.Do(func() {
		ctx := context.Background()

		rdContainer, err = redisC.Run(ctx,
			"redis:7-alpine",
		)

		if err != nil {
			err = fmt.Errorf("failed to start container: %w", err)
			return
		}

		connStr, err := rdContainer.ConnectionString(ctx)
		if err != nil {
			err = fmt.Errorf("failed to get host: %w", err)
			return
		}

		parsedURL, err := url.Parse(connStr)
		if err != nil {
			err = fmt.Errorf("failed to parse connection string: %w", err)
			return
		}

		rdHost = parsedURL.Host
	})

	rdTest, err = redisRepo.NewRedisClient(context.Background(), redisRepo.RedisConfig{
		Addr:     rdHost,
		Password: "",
		DB:       0,
	})

	rdTestClient = rdTest.Client()
	queueTestRepo = repository.NewQueue(rdTestClient)
	return err
}

func cleanupTestRD(t *testing.T) {
	ctx := context.Background()
	err := rdTestClient.FlushDB(ctx).Err()
	require.NoError(t, err)
}

func createTestLocationCheck(id int, userID string) *domain.LocationCheck {
	return &domain.LocationCheck{
		ID:           id,
		UserID:       userID,
		CheckedAt:    time.Now(),
		Lat:          55.7558,
		Long:         37.6173,
		InDangerZone: true,
		NearestID:    func() *int { v := 1; return &v }(),
	}
}

func TestWebhookWorker_Success(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	if err := setup(); err != nil {
		t.Fatalf("failed to setup: %v", err)
	}
	defer cleanupTestRD(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	webhookReceived := make(chan *domain.LocationCheck, 1)

	server := createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var check domain.LocationCheck
		err := json.NewDecoder(r.Body).Decode(&check)
		require.NoError(t, err)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "webhook received"})

		select {
		case webhookReceived <- &check:
		default:
		}
	})

	logger := createTestLogger()
	worker := NewWebhookWorker(queueTestRepo, server.URL, logger)

	check := createTestLocationCheck(1, "colorvax")
	err := queueTestRepo.Enqueue(ctx, check)
	require.NoError(t, err)

	queueLenBefore, err := rdTestClient.LLen(ctx, "webhook:queue").Result()
	require.NoError(t, err)
	require.Equal(t, int64(1), queueLenBefore)

	done := worker.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	var received *domain.LocationCheck
	select {
	case received = <-webhookReceived:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for webhook to be received")
	}

	require.NotNil(t, received)
	require.Equal(t, 1, received.ID)
	require.Equal(t, "colorvax", received.UserID)

	queueLen, err := rdTestClient.LLen(ctx, "webhook:queue").Result()
	require.NoError(t, err)
	require.Equal(t, int64(0), queueLen)

	dlqLen, err := rdTestClient.LLen(ctx, "webhook:dlq").Result()
	require.NoError(t, err)
	require.Equal(t, int64(0), dlqLen)

	cancel()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("timeout waiting for worker to stop")
	}
}

func TestWebhookWorker_RetryOnInternalServerError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	if err := setup(); err != nil {
		t.Fatalf("failed to setup: %v", err)
	}
	defer cleanupTestRD(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	attemptCount := 0
	webhookReceived := make(chan *domain.LocationCheck, 1)

	server := createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		attemptCount++

		if attemptCount <= 2 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		var check domain.LocationCheck
		err := json.NewDecoder(r.Body).Decode(&check)
		require.NoError(t, err)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "webhook received"})

		select {
		case webhookReceived <- &check:
		default:
		}
	})

	logger := createTestLogger()
	worker := NewWebhookWorker(queueTestRepo, server.URL, logger)

	check := createTestLocationCheck(1, "colorvax")
	err := queueTestRepo.Enqueue(ctx, check)
	require.NoError(t, err)

	queueLenBefore, err := rdTestClient.LLen(ctx, "webhook:queue").Result()
	require.NoError(t, err)
	require.Equal(t, int64(1), queueLenBefore)

	done := worker.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	var received *domain.LocationCheck
	select {
	case received = <-webhookReceived:
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for webhook to be received after retry")
	}

	require.GreaterOrEqual(t, attemptCount, 2)

	require.NotNil(t, received)
	require.Equal(t, 1, received.ID)
	require.Equal(t, "colorvax", received.UserID)

	queueLen, err := rdTestClient.LLen(ctx, "webhook:queue").Result()
	require.NoError(t, err)
	require.Equal(t, int64(0), queueLen)

	dlqLen, err := rdTestClient.LLen(ctx, "webhook:dlq").Result()
	require.NoError(t, err)
	require.Equal(t, int64(0), dlqLen)

	cancel()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("timeout waiting for worker to stop")
	}
}

func TestWebhook_MoveToDLQAfterMaxRetries(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	if err := setup(); err != nil {
		t.Fatalf("failed to setup: %v", err)
	}
	defer cleanupTestRD(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	logger := createTestLogger()
	worker := NewWebhookWorker(queueTestRepo, server.URL, logger)

	check := createTestLocationCheck(1, "colorvax")
	err := queueTestRepo.Enqueue(ctx, check)
	require.NoError(t, err)

	queueLenBefore, err := rdTestClient.LLen(ctx, "webhook:queue").Result()
	require.NoError(t, err)
	require.Equal(t, int64(1), queueLenBefore)

	done := worker.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	<-time.After(7 * time.Second)

	queueLen, err := rdTestClient.LLen(ctx, "webhook:queue").Result()
	require.NoError(t, err)
	require.Equal(t, int64(0), queueLen)

	queueDelayedLen, err := rdTestClient.ZCard(ctx, "webhook:delayed").Result()
	require.NoError(t, err)
	require.Equal(t, int64(0), queueDelayedLen)

	dlqLen, err := rdTestClient.LLen(ctx, "webhook:dlq").Result()
	require.NoError(t, err)
	require.Equal(t, int64(1), dlqLen)

	dlqData, err := rdTestClient.LPop(ctx, "webhook:dlq").Result()
	require.NoError(t, err)
	require.NotEmpty(t, dlqData)

	var task repository.WebhookTask
	err = json.Unmarshal([]byte(dlqData), &task)
	require.NoError(t, err)

	require.Equal(t, maxRetries, task.Attempt)
	require.NotEmpty(t, task.LastError)
	require.Contains(t, task.LastError, "500")

	cancel()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("timeout waiting for worker to stop")
	}
}

func TestWebhook_NotRetryableError(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	if err := setup(); err != nil {
		t.Fatalf("failed to setup: %v", err)
	}
	defer cleanupTestRD(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	server := createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	logger := createTestLogger()
	worker := NewWebhookWorker(queueTestRepo, server.URL, logger)

	check := createTestLocationCheck(1, "colorvax")
	err := queueTestRepo.Enqueue(ctx, check)
	require.NoError(t, err)

	queueLenBefore, err := rdTestClient.LLen(ctx, "webhook:queue").Result()
	require.NoError(t, err)
	require.Equal(t, int64(1), queueLenBefore)

	done := worker.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	<-time.After(2 * time.Second)

	queueLen, err := rdTestClient.LLen(ctx, "webhook:queue").Result()
	require.NoError(t, err)
	require.Equal(t, int64(0), queueLen)

	queueDelayedLen, err := rdTestClient.ZCard(ctx, "webhook:delayed").Result()
	require.NoError(t, err)
	require.Equal(t, int64(0), queueDelayedLen)

	dlqLen, err := rdTestClient.LLen(ctx, "webhook:dlq").Result()
	require.NoError(t, err)
	require.Equal(t, int64(1), dlqLen)

	dlqData, err := rdTestClient.LPop(ctx, "webhook:dlq").Result()
	require.NoError(t, err)
	require.NotEmpty(t, dlqData)

	var task repository.WebhookTask
	err = json.Unmarshal([]byte(dlqData), &task)
	require.NoError(t, err)

	require.Equal(t, 1, task.Attempt)
	require.NotEmpty(t, task.LastError)
	require.Contains(t, task.LastError, "400")

	cancel()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("timeout waiting for worker to stop")
	}
}

func TestWebhook_DelayedTaskProcessing(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	if err := setup(); err != nil {
		t.Fatalf("failed to setup: %v", err)
	}
	defer cleanupTestRD(t)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	webhookReceived := make(chan *domain.LocationCheck, 1)

	server := createTestServer(t, func(w http.ResponseWriter, r *http.Request) {
		var check domain.LocationCheck
		err := json.NewDecoder(r.Body).Decode(&check)
		require.NoError(t, err)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"message": "webhook received"})

		select {
		case webhookReceived <- &check:
		default:
		}
	})

	logger := createTestLogger()
	worker := NewWebhookWorker(queueTestRepo, server.URL, logger)

	task := &repository.WebhookTask{
		LocationCheck: createTestLocationCheck(1, "colorvax"),
		Attempt:       1,
		FirstAttempt:  time.Now(),
	}

	err := queueTestRepo.EnqueueWithDelay(ctx, task, 100*time.Millisecond)
	require.NoError(t, err)

	queueLenBefore, err := rdTestClient.ZCard(ctx, "webhook:delayed").Result()
	require.NoError(t, err)
	require.Equal(t, int64(1), queueLenBefore)

	done := worker.Start(ctx)
	time.Sleep(50 * time.Millisecond)

	var received *domain.LocationCheck
	select {
	case received = <-webhookReceived:
	case <-time.After(2 * time.Second):
		t.Fatal("timeout waiting for webhook to be received")
	}

	require.NotNil(t, received)
	require.Equal(t, 1, received.ID)
	require.Equal(t, "colorvax", received.UserID)

	queueLen, err := rdTestClient.LLen(ctx, "webhook:queue").Result()
	require.NoError(t, err)
	require.Equal(t, int64(0), queueLen)

	queueLenBefore, err = rdTestClient.ZCard(ctx, "webhook:delayed").Result()
	require.NoError(t, err)
	require.Equal(t, int64(0), queueLenBefore)

	dlqLen, err := rdTestClient.LLen(ctx, "webhook:dlq").Result()
	require.NoError(t, err)
	require.Equal(t, int64(0), dlqLen)

	cancel()

	select {
	case <-done:
	case <-time.After(3 * time.Second):
		t.Fatalf("timeout waiting for worker to stop")
	}
}

func TestWebhook_NetworkError(t *testing.T) {}
