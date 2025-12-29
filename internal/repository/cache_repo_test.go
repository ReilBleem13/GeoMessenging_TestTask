package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"sync"
	"testing"

	"red_collar/internal/domain"
	redisRepo "red_collar/internal/repository/redis"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	redisC "github.com/testcontainers/testcontainers-go/modules/redis"
)

var (
	rdContainer *redisC.RedisContainer
	rdOnce      sync.Once
	rdHost      string

	testRDClient *redisRepo.RedisClient
	testRD       *redis.Client

	testRDRepo    *CacheRepository
	testQueueRepo *Queue
)

func setupTestRD() error {
	var err error

	rdOnce.Do(func() {
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

	testRDClient, err = redisRepo.NewRedisClient(context.Background(), redisRepo.RedisConfig{
		Addr:     rdHost,
		Password: "",
		DB:       0,
	})

	testRD = testRDClient.Client()
	testRDRepo = NewCacheRepository(testRD)
	return err
}

func setupQueue() {
	setupTestRD()
	testQueueRepo = NewQueue(testRD)
}

func cleanupTestRD(t *testing.T) {
	ctx := context.Background()
	err := testRD.FlushDB(ctx).Err()
	require.NoError(t, err)
}

func TestCacheRepository_Save(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	if testRD == nil {
		setupTestRD()
	}

	ctx := context.Background()

	tests := []struct {
		name     string
		data     func() []byte
		key      string
		validate func(t *testing.T, key string)
	}{
		{
			name: "success",
			data: func() []byte {
				incident := &domain.Incident{
					ID:    1,
					Title: "colorvax",
				}
				data, err := json.Marshal(incident)
				require.NoError(t, err)
				return data
			},
			key: "incidentID:1",
			validate: func(t *testing.T, key string) {
				res, err := testRDRepo.Get(ctx, key)
				require.NoError(t, err)

				var incident *domain.Incident
				err = json.Unmarshal(res, &incident)
				require.NoError(t, err)

				require.Equal(t, 1, incident.ID)
				require.Equal(t, "colorvax", incident.Title)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanupTestRD(t)

			err := testRDRepo.Save(ctx, tt.data(), tt.key)
			require.NoError(t, err)

			if tt.validate != nil {
				tt.validate(t, tt.key)
			}
		})
	}
}

func TestCacheRepository_Get(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	if testRD == nil {
		setupTestRD()
	}

	ctx := context.Background()

	tests := []struct {
		name     string
		key      string
		setup    func(t *testing.T, key string)
		validate func(t *testing.T, res []byte)
	}{
		{
			name: "success",
			key:  "incidentID:1",
			setup: func(t *testing.T, key string) {
				incident := &domain.Incident{
					ID:    1,
					Title: "colorvax",
				}
				data, err := json.Marshal(incident)
				require.NoError(t, err)

				err = testRDRepo.Save(ctx, data, key)
				require.NoError(t, err)
			},
			validate: func(t *testing.T, res []byte) {
				var incident *domain.Incident
				err := json.Unmarshal(res, &incident)
				require.NoError(t, err)

				require.Equal(t, 1, incident.ID)
				require.Equal(t, "colorvax", incident.Title)
			},
		},
		{
			name: "success - but no data",
			key:  "incidentID:1",
			validate: func(t *testing.T, res []byte) {
				require.Nil(t, res)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanupTestRD(t)

			if tt.setup != nil {
				tt.setup(t, tt.key)
			}

			data, err := testRDRepo.Get(ctx, tt.key)
			require.NoError(t, err)

			if tt.validate != nil {
				tt.validate(t, data)
			}
		})
	}
}

func TestCacheRepository_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	if testRD == nil {
		setupTestRD()
	}

	ctx := context.Background()

	tests := []struct {
		name     string
		key      string
		setup    func(t *testing.T, key string)
		validate func(t *testing.T, key string, deleted bool)
	}{
		{
			name: "success",
			key:  "incidentID:1",
			setup: func(t *testing.T, key string) {
				incident := &domain.Incident{
					ID:    1,
					Title: "colorvax",
				}
				data, err := json.Marshal(incident)
				require.NoError(t, err)

				err = testRDRepo.Save(ctx, data, key)
				require.NoError(t, err)

				data, err = testRDRepo.Get(ctx, key)
				require.NotNil(t, data)
				require.NoError(t, err)
			},
			validate: func(t *testing.T, key string, deleted bool) {
				require.True(t, deleted)

				data, err := testRDRepo.Get(ctx, key)
				require.Nil(t, data)
				require.NoError(t, err)
			},
		},
		{
			name: "success - no data",
			key:  "incidentID:1",
			validate: func(t *testing.T, key string, deleted bool) {
				require.False(t, deleted)

				data, err := testRDRepo.Get(ctx, key)
				require.Nil(t, data)
				require.NoError(t, err)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanupTestRD(t)

			if tt.setup != nil {
				tt.setup(t, tt.key)
			}

			deleted, err := testRDRepo.Delete(ctx, tt.key)
			require.NoError(t, err)

			if tt.validate != nil {
				tt.validate(t, tt.key, deleted)
			}
		})
	}
}
