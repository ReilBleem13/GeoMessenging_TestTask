package worker

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"red_collar/internal/domain"
	"red_collar/internal/repository/redis"
	"red_collar/internal/service"
	"time"

	"github.com/theartofdevel/logging"
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

	for {
		select {
		case <-ctx.Done():
			w.logger.Info("webhook worker stopping...")
			return
		default:
			check, err := w.queue.Dequeue(ctx)
			if err != nil {
				w.logger.Error("failed to dequeue location check", logging.ErrAttr(err))
				time.Sleep(1 * time.Second)
				continue
			}

			if check == nil {
				continue
			}

			if err := w.sendWebhook(ctx, check); err != nil {
				w.logger.Error("failed to send webhook",
					logging.StringAttr("user_id", check.UserID),
					logging.ErrAttr(err),
				)
				// подумать над ретраями и dlq
				continue
			}

			w.logger.Info("webhook sent successfully",
				logging.StringAttr("user_id", check.UserID),
				logging.IntAttr("check_id", check.ID),
			)
		}
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
		return fmt.Errorf("webhook returned not OK: %d", resp.StatusCode)
	}
	return nil
}
