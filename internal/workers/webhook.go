package worker

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"red_collar/internal/domain"
	"red_collar/internal/repository/redis"
	"red_collar/internal/service"
	"time"

	"github.com/theartofdevel/logging"
)

const (
	maxRetries = 3
	baseDelay  = 1 * time.Second
)

type WebhookWorker struct {
	queue      *redis.Queue
	webhookURL string
	logger     service.LoggerInterfaces
	client     *http.Client
}

func NewWebhookWorker(queue *redis.Queue, webhookURL string, logger service.LoggerInterfaces) *WebhookWorker {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	return &WebhookWorker{
		queue:      queue,
		webhookURL: webhookURL,
		logger:     logger,
		client: &http.Client{
			Timeout:   10 * time.Second,
			Transport: tr,
		},
	}
}

func (w *WebhookWorker) Start(ctx context.Context) {
	w.logger.Info("webhook worker started", logging.StringAttr("webhook_url", w.webhookURL))

	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := w.queue.ProcessDelayedTasks(ctx); err != nil {
					w.logger.Error("failed to process delayed tasks", logging.ErrAttr(err))
				}
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("webhook worker stopping...")
			return
		default:
			data, err := w.queue.Dequeue(ctx)
			if err != nil {
				w.logger.Error("failed to dequeue location check", logging.ErrAttr(err))
				time.Sleep(1 * time.Second)
				continue
			}

			if data == nil {
				continue
			}

			if data.Attempt > 0 {
				w.logger.Info("retrying webhook",
					logging.IntAttr("attempt", data.Attempt),
					logging.StringAttr("user_id", data.LocationCheck.UserID),
				)
			}

			if err := w.sendWebhook(ctx, data.LocationCheck); err != nil {
				w.handleWebhookError(ctx, data, err)
				continue
			}

			w.logger.Info("webhook sent successfully",
				logging.StringAttr("user_id", data.LocationCheck.UserID),
				logging.IntAttr("check_id", data.LocationCheck.ID),
			)
		}
	}
}

func (w *WebhookWorker) handleWebhookError(ctx context.Context, task *redis.WebhookTask, err error) {
	task.Attempt++
	task.LastError = err.Error()

	if w.isRetryable(err) && task.Attempt < maxRetries {
		if err := w.queue.EnqueueWithDelay(ctx, task, w.calculateBackoff(task.Attempt)); err != nil {
			w.logger.Error("failed to enqueue retry",
				logging.StringAttr("user_id", task.LocationCheck.UserID),
				logging.IntAttr("attempt", task.Attempt),
				logging.ErrAttr(err),
			)
			w.sendToDLQ(ctx, task)
		} else {
			w.logger.Info("webhook task scheduled for retry",
				logging.StringAttr("user_id", task.LocationCheck.UserID),
				logging.IntAttr("attempt", task.Attempt),
			)
		}
	} else {
		w.sendToDLQ(ctx, task)
	}

}

func (w *WebhookWorker) sendToDLQ(ctx context.Context, task *redis.WebhookTask) {
	if err := w.queue.EnqueueDLQ(ctx, task); err != nil {
		w.logger.Error("failed to send task to DLQ",
			logging.StringAttr("user_id", task.LocationCheck.UserID),
			logging.ErrAttr(err),
		)
	} else {
		w.logger.Warn("webhook task moved to DLQ",
			logging.StringAttr("user_id", task.LocationCheck.UserID),
			logging.IntAttr("check_id", task.LocationCheck.ID),
			logging.IntAttr("final_attempt", task.Attempt),
			logging.StringAttr("last_error", task.LastError),
		)
	}
}

func (w *WebhookWorker) sendWebhook(ctx context.Context, check *domain.LocationCheck) error {
	data, err := json.Marshal(check)
	if err != nil {
		return fmt.Errorf("failed to marshal location check: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.webhookURL, bytes.NewBuffer(data))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := w.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return &httpError{
			statusCode: resp.StatusCode,
			message:    fmt.Sprintf("webhook returned status: %d", resp.StatusCode),
		}
	}
	return nil
}

func (w *WebhookWorker) isRetryable(err error) bool {
	if err == nil {
		return false
	}

	var httpErr *httpError
	if errors.As(err, &httpErr) {
		if httpErr.statusCode >= 500 && httpErr.statusCode < 600 {
			return true
		}

		if httpErr.statusCode == 429 {
			return true
		}
		return false
	}
	return true
}

func (w *WebhookWorker) calculateBackoff(attempt int) time.Duration {
	delay := baseDelay * time.Duration(1<<uint(attempt-1))

	if delay > 30*time.Second {
		delay = 30 * time.Second
	}
	return delay
}

type httpError struct {
	statusCode int
	message    string
}

func (e *httpError) Error() string {
	return e.message
}
