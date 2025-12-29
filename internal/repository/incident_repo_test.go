package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"red_collar/internal/domain"
	"red_collar/internal/repository/database"
	"sort"
	"sync"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var (
	pgContainer *postgres.PostgresContainer
	pgConnStr   string
	pgOnce      sync.Once

	testDB       *sqlx.DB
	testDBClient *database.PostgresClient

	testRepo     *IncidentRepository
	testRepoCoor *CoordinatesRepository
)

// настройка
func setupTestDBContainer() (string, error) {
	var err error

	pgOnce.Do(func() {
		ctx := context.Background()

		pgContainer, err = postgres.Run(ctx,
			"postgis/postgis:15-3.4",
			postgres.WithDatabase("testdb"),
			postgres.WithUsername("testuser"),
			postgres.WithPassword("testpass"),
			testcontainers.WithWaitStrategy(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2).
					WithStartupTimeout(60*time.Second),
			),
		)
		if err != nil {
			err = fmt.Errorf("failed to start container: %w", err)
			return
		}

		pgConnStr, err = pgContainer.ConnectionString(ctx, "sslmode=disable")
		if err != nil {
			err = fmt.Errorf("failed to get connection string: %w", err)
			return
		}

		if err = verifyPostGIS(ctx, pgConnStr); err != nil {
			err = fmt.Errorf("postgis verification failed: %w", err)
			return
		}
	})
	return pgConnStr, err
}

func verifyPostGIS(ctx context.Context, dsn string) error {
	db, err := sqlx.ConnectContext(ctx, "postgres", dsn)
	if err != nil {
		return fmt.Errorf("failed to connect for postgis check: %w", err)
	}
	defer db.Close()

	var version string
	err = db.GetContext(ctx, &version, "SELECT PostGIS_version()")
	if err != nil {
		return fmt.Errorf("postgis extension not available: %w", err)
	}

	if version == "" {
		return fmt.Errorf("postgis version is empty")
	}
	return nil
}

func setupTestDB(t *testing.T) {
	dsn, err := setupTestDBContainer()
	require.NoError(t, err)

	testDBClient, err = database.NewPostgresClient(context.Background(), dsn)
	require.NoError(t, err)

	testDB = testDBClient.Client()

	goose.SetDialect("postgres")
	migrationPath := "../../migrations"

	err = goose.Up(testDB.DB, migrationPath)
	require.NoError(t, err)

	testRepo = NewIncidentRepository(testDB)
	testRepoCoor = NewCoordinatesRepository(testDB)
}

func cleanupTestDB(t *testing.T) {
	_, err := testDB.Exec("TRUNCATE TABLE location_checks, incidents RESTART IDENTITY CASCADE")
	require.NoError(t, err)
}

func TestMain(m *testing.M) {
	code := m.Run()

	if testDBClient != nil {
		testDBClient.Close()
	}
	os.Exit(code)
}

// тесты
func TestIncidentRepository_Create(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	if testDB == nil {
		setupTestDB(t)
	}

	ctx := context.Background()

	tests := []struct {
		name     string
		incident *domain.Incident
		setup    func(t *testing.T)
		wantErr  bool
		errType  func(err error) bool
		validate func(t *testing.T, incident *domain.Incident)
	}{
		{
			name: "success",
			incident: &domain.Incident{
				Title:       "Incident",
				Description: "Description",
				Lat:         50.0,
				Long:        30.0,
				Radius:      100,
				Active:      true,
			},
			validate: func(t *testing.T, incident *domain.Incident) {
				require.NotZero(t, incident.ID)
				require.Equal(t, "Incident", incident.Title)
				require.Equal(t, "Description", incident.Description)
				require.Equal(t, 50.0, incident.Lat)
				require.Equal(t, 30.0, incident.Long)
				require.Equal(t, 100, incident.Radius)
				require.True(t, incident.Active)
				require.WithinDuration(t, time.Now(), incident.CreatedAt, 10*time.Second)
				require.WithinDuration(t, time.Now(), incident.UpdatedAt, 10*time.Second)
			},
		},
		{
			name: "error - duplicate title",
			incident: &domain.Incident{
				Title:       "Incident",
				Description: "Description",
				Lat:         50.0,
				Long:        30.0,
				Radius:      100,
				Active:      true,
			},
			setup: func(t *testing.T) {
				incident := &domain.Incident{
					Title:       "Incident",
					Description: "Description",
					Lat:         50.0,
					Long:        30.0,
					Radius:      100,
					Active:      true,
				}
				err := testRepo.Create(ctx, incident)
				require.NoError(t, err)
			},
			wantErr: true,
			errType: func(err error) bool {
				var appErr *domain.AppError
				if !errors.As(err, &appErr) {
					return false
				}
				return appErr.Code == domain.CodeAlreadyExists
			},
			validate: func(t *testing.T, incident *domain.Incident) {
				require.Zero(t, incident.ID)
				require.Zero(t, incident.CreatedAt)
				require.Zero(t, incident.UpdatedAt)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanupTestDB(t)

			if tt.setup != nil {
				tt.setup(t)
			}

			incident := &domain.Incident{
				Title:       tt.incident.Title,
				Description: tt.incident.Description,
				Lat:         tt.incident.Lat,
				Long:        tt.incident.Long,
				Radius:      tt.incident.Radius,
				Active:      tt.incident.Active,
			}

			err := testRepo.Create(ctx, incident)

			if tt.wantErr {
				require.Error(t, err, "expected error but got none")
				if tt.errType != nil {
					require.True(t, tt.errType(err), "wrong error type")
				}
			} else {
				require.NoError(t, err, "unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, incident)
			}
		})
	}
}

func TestIncidentRepository_GetByID(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	if testDB == nil {
		setupTestDB(t)
	}

	ctx := context.Background()

	tests := []struct {
		name     string
		id       int
		setup    func(t *testing.T)
		wantErr  bool
		errType  func(err error) bool
		validate func(t *testing.T, incident *domain.Incident)
	}{
		{
			name: "success",
			id:   1,
			setup: func(t *testing.T) {
				incident := &domain.Incident{
					Title:       "Incident",
					Description: "Description",
					Lat:         50.0,
					Long:        30.0,
					Radius:      100,
					Active:      true,
				}
				err := testRepo.Create(ctx, incident)
				require.NoError(t, err)
			},
			validate: func(t *testing.T, incident *domain.Incident) {
				require.NotZero(t, incident.ID)
				require.Equal(t, "Incident", incident.Title)
				require.Equal(t, "Description", incident.Description)
				require.Equal(t, 50.0, incident.Lat)
				require.Equal(t, 30.0, incident.Long)
				require.Equal(t, 100, incident.Radius)
				require.True(t, incident.Active)
				require.WithinDuration(t, time.Now(), incident.CreatedAt, 10*time.Second)
				require.WithinDuration(t, time.Now(), incident.UpdatedAt, 10*time.Second)
			},
		},
		{
			name:    "error - incident not found",
			id:      1,
			wantErr: true,
			errType: func(err error) bool {
				var appErr *domain.AppError
				if !errors.As(err, &appErr) {
					return false
				}
				return appErr.Code == domain.CodeNotFound
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanupTestDB(t)

			if tt.setup != nil {
				tt.setup(t)
			}

			incident, err := testRepo.GetByID(ctx, tt.id)

			if tt.wantErr {
				require.Error(t, err, "expected error but got none")
				if tt.errType != nil {
					require.True(t, tt.errType(err), "wrong error type")
				}
			} else {
				require.NoError(t, err, "unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, incident)
			}
		})
	}
}

func TestIncidentRepository_Paginate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	if testDB == nil {
		setupTestDB(t)
	}

	tests := []struct {
		name          string
		limit, offset int
		setup         func(t *testing.T)
		wantErr       bool
		errType       func(err error) bool
		validate      func(t *testing.T, incidents []domain.Incident, total int)
	}{
		{
			name:   "success",
			limit:  5,
			offset: 0,
			setup: func(t *testing.T) {
				for i := 0; i < 5; i++ {
					incident := &domain.Incident{
						Title:       fmt.Sprintf("Incident-#%d", i),
						Description: "Description",
						Lat:         50.0,
						Long:        30.0,
						Radius:      100,
						Active:      true,
					}
					err := testRepo.Create(ctx, incident)
					require.NoError(t, err)
				}
			},
			validate: func(t *testing.T, incidents []domain.Incident, total int) {
				require.Equal(t, 5, total)
				require.Equal(t, 5, len(incidents))

				sort.Slice(incidents, func(i, j int) bool {
					return incidents[i].ID < incidents[j].ID
				})

				for i := 0; i < 5; i++ {
					require.Equal(t, incidents[i].Title, fmt.Sprintf("Incident-#%d", i))
					require.Equal(t, "Description", incidents[i].Description)
					require.Equal(t, 50.0, incidents[i].Lat)
					require.Equal(t, 30.0, incidents[i].Long)
					require.Equal(t, 100, incidents[i].Radius)
					require.True(t, incidents[i].Active)
					require.NotZero(t, incidents[i].ID)
					require.WithinDuration(t, time.Now(), incidents[i].CreatedAt, 10*time.Second)
					require.WithinDuration(t, time.Now(), incidents[i].UpdatedAt, 10*time.Second)
				}
			},
		},
		{
			name:   "success",
			limit:  5,
			offset: 0,
			setup: func(t *testing.T) {
				for i := 0; i < 5; i++ {
					incident := &domain.Incident{
						Title:       fmt.Sprintf("Incident-#%d", i),
						Description: "Description",
						Lat:         50.0,
						Long:        30.0,
						Radius:      100,
						Active:      true,
					}
					err := testRepo.Create(ctx, incident)
					require.NoError(t, err)
				}
			},
			validate: func(t *testing.T, incidents []domain.Incident, total int) {
				require.Equal(t, 5, total)
				require.Equal(t, 5, len(incidents))

				sort.Slice(incidents, func(i, j int) bool {
					return incidents[i].ID < incidents[j].ID
				})

				for i := 0; i < 5; i++ {
					require.Equal(t, fmt.Sprintf("Incident-#%d", i), incidents[i].Title)
					require.Equal(t, "Description", incidents[i].Description)
					require.Equal(t, 50.0, incidents[i].Lat)
					require.Equal(t, 30.0, incidents[i].Long)
					require.Equal(t, 100, incidents[i].Radius)
					require.True(t, incidents[i].Active)
					require.NotZero(t, incidents[i].ID)
					require.WithinDuration(t, time.Now(), incidents[i].CreatedAt, 10*time.Second)
					require.WithinDuration(t, time.Now(), incidents[i].UpdatedAt, 10*time.Second)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanupTestDB(t)

			if tt.setup != nil {
				tt.setup(t)
			}

			incidents, total, err := testRepo.Paginate(ctx, tt.limit, tt.offset)

			if tt.wantErr {
				require.Error(t, err, "expected error but got none")
				if tt.errType != nil {
					require.True(t, tt.errType(err), "wrong error type")
				}
			} else {
				require.NoError(t, err, "unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, incidents, total)
			}
		})
	}
}

func TestIncidentRepository_Delete(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	if testDB == nil {
		setupTestDB(t)
	}

	tests := []struct {
		name     string
		id       int
		setup    func(t *testing.T)
		wantErr  bool
		errType  func(err error) bool
		validate func(t *testing.T)
	}{
		{
			name: "success",
			id:   1,
			setup: func(t *testing.T) {
				incident := &domain.Incident{
					Title:       "Incident",
					Description: "Description",
					Lat:         50.0,
					Long:        30.0,
					Radius:      100,
					Active:      true,
				}
				err := testRepo.Create(ctx, incident)
				require.NoError(t, err)
			},
			validate: func(t *testing.T) {
				incident, err := testRepo.GetByID(ctx, 1)
				require.Error(t, err)
				require.Nil(t, incident)

				var appErr *domain.AppError
				require.True(t, errors.As(err, &appErr))
				require.True(t, appErr.Code == domain.CodeNotFound)
			},
		},
		{
			name:    "error - incident not found",
			id:      1,
			wantErr: true,
			errType: func(err error) bool {
				var appErr *domain.AppError
				if !errors.As(err, &appErr) {
					return false
				}
				return appErr.Code == domain.CodeNotFound
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanupTestDB(t)

			if tt.setup != nil {
				tt.setup(t)
			}

			err := testRepo.Delete(ctx, tt.id)

			if tt.wantErr {
				require.Error(t, err, "expected error but got none")
				if tt.errType != nil {
					require.True(t, tt.errType(err), "wrong error type")
				}
			} else {
				require.NoError(t, err, "unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t)
			}
		})
	}
}

func TestIncidentRepository_FullUpdate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}

	ctx := context.Background()

	if testDB == nil {
		setupTestDB(t)
	}

	tests := []struct {
		name     string
		incident *domain.Incident
		setup    func(t *testing.T)
		wantErr  bool
		errType  func(err error) bool
		validate func(t *testing.T, incident *domain.Incident)
	}{
		{
			name: "success",
			incident: &domain.Incident{
				ID:          1,
				Title:       "Incident-Updated",
				Description: "Description",
				Lat:         75.0,
				Long:        75.0,
				Radius:      75,
				Active:      true,
			},
			setup: func(t *testing.T) {
				incident := &domain.Incident{
					Title:       "Incident",
					Description: "Description",
					Lat:         50.0,
					Long:        50.0,
					Radius:      50,
					Active:      true,
				}
				err := testRepo.Create(ctx, incident)
				require.NoError(t, err)
			},
			validate: func(t *testing.T, incident *domain.Incident) {
				require.Equal(t, 1, incident.ID)
				require.Equal(t, "Incident-Updated", incident.Title)
				require.Equal(t, "Description", incident.Description)
				require.Equal(t, 75.0, incident.Lat)
				require.Equal(t, 75.0, incident.Long)
				require.Equal(t, 75, incident.Radius)
				require.True(t, incident.Active)
				require.WithinDuration(t, time.Now(), incident.CreatedAt, 10*time.Second)
				require.WithinDuration(t, time.Now(), incident.UpdatedAt, 10*time.Second)
				require.True(t, incident.UpdatedAt.After(incident.CreatedAt))
			},
		},
		{
			name: "error - incident not found",
			incident: &domain.Incident{
				ID:          1,
				Title:       "Incident-Updated",
				Description: "Description",
				Lat:         75.0,
				Long:        75.0,
				Radius:      75,
				Active:      true,
			},
			wantErr: true,
			errType: func(err error) bool {
				var appErr *domain.AppError
				if !errors.As(err, &appErr) {
					return false
				}
				return appErr.Code == domain.CodeNotFound
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanupTestDB(t)

			if tt.setup != nil {
				tt.setup(t)
			}

			incident := &domain.Incident{
				ID:          tt.incident.ID,
				Title:       tt.incident.Title,
				Description: tt.incident.Description,
				Lat:         tt.incident.Lat,
				Long:        tt.incident.Long,
				Radius:      tt.incident.Radius,
				Active:      tt.incident.Active,
			}

			err := testRepo.FullUpdate(ctx, incident)

			if tt.wantErr {
				require.Error(t, err, "expected error but got none")
				if tt.errType != nil {
					require.True(t, tt.errType(err), "wrong error type")
				}
			} else {
				require.NoError(t, err, "unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, incident)
			}
		})
	}
}
