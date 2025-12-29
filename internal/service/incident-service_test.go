package service

import (
	"context"
	"encoding/json"
	"errors"
	"red_collar/internal/domain"
	"testing"
)

func TestService_CreateIncident(t *testing.T) {
	tests := []struct {
		name           string
		input          *CreateIncidentRequestInput
		incidents      func() *mockIncidentsRepository
		wantErr        bool
		validateResult func(t *testing.T, result *domain.Incident)
		validateLogs   func(t *testing.T, logger *mockLogger)
	}{
		{
			name:  "validation error - title is empty",
			input: &CreateIncidentRequestInput{},
			incidents: func() *mockIncidentsRepository {
				return &mockIncidentsRepository{}
			},
			wantErr: true,
			validateLogs: func(t *testing.T, logger *mockLogger) {
				errorsLogs := logger.GetErrorLogs()
				if len(errorsLogs) != 1 {
					t.Error("expected 1 error log, got none")
					return
				}

				expectedMsg := "create incident request validation failed"
				if errorsLogs[0].msg != expectedMsg {
					t.Errorf("expected error message '%s' not found", expectedMsg)
				}
			},
		},
		{
			name: "validation error - lat is invalid",
			input: &CreateIncidentRequestInput{
				Title:  "Color",
				Lat:    -120,
				Long:   100,
				Radius: 55,
			},
			incidents: func() *mockIncidentsRepository {
				return &mockIncidentsRepository{}
			},
			wantErr: true,
			validateLogs: func(t *testing.T, logger *mockLogger) {
				errorsLogs := logger.GetErrorLogs()
				if len(errorsLogs) != 1 {
					t.Error("expected 1 error log, got none")
					return
				}

				expectedMsg := "create incident request validation failed"
				if errorsLogs[0].msg != expectedMsg {
					t.Errorf("expected error message '%s' not found", expectedMsg)
				}
			},
		},
		{
			name: "validation error - radius is invalid",
			input: &CreateIncidentRequestInput{
				Title:  "Color",
				Lat:    -20,
				Long:   100,
				Radius: 10,
			},
			incidents: func() *mockIncidentsRepository {
				return &mockIncidentsRepository{}
			},
			wantErr: true,
			validateLogs: func(t *testing.T, logger *mockLogger) {
				errorsLogs := logger.GetErrorLogs()
				if len(errorsLogs) != 1 {
					t.Error("expected 1 error log, got none")
					return
				}

				expectedMsg := "create incident request validation failed"
				if errorsLogs[0].msg != expectedMsg {
					t.Errorf("expected error message '%s' not found", expectedMsg)
				}
			},
		},
		{
			name: "success - active & description is empty",
			input: &CreateIncidentRequestInput{
				Title:  "Color",
				Lat:    -20,
				Long:   100,
				Radius: 55,
			},
			incidents: func() *mockIncidentsRepository {
				return &mockIncidentsRepository{
					createFunc: func(ctx context.Context, incident *domain.Incident) error {
						return nil
					},
				}
			},
			validateResult: func(t *testing.T, result *domain.Incident) {
				if result.Active != true {
					t.Error("expected true, got false")
					return
				}

				if result.Description != "without description" {
					t.Errorf("expected 'without description', got '%s'", result.Description)
					return
				}
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				infoLogs := logger.GetInfoLogs()
				if len(infoLogs) != 2 {
					t.Error("expected 2 info logs, got none")
					return
				}
			},
		},
		{
			name: "success",
			input: &CreateIncidentRequestInput{
				Title:       "Color",
				Description: ptr("something"),
				Lat:         -20,
				Long:        100,
				Radius:      55,
				Active:      ptr(true),
			},
			incidents: func() *mockIncidentsRepository {
				return &mockIncidentsRepository{
					createFunc: func(ctx context.Context, incident *domain.Incident) error {
						return nil
					},
				}
			},
			validateResult: func(t *testing.T, result *domain.Incident) {
				if result.Active != true {
					t.Error("expected true, got false")
					return
				}

				if result.Description != "something" {
					t.Errorf("expected 'something', got '%s'", result.Description)
					return
				}

				if result.Title != "Color" {
					t.Errorf("expected title 'Color', got '%s'", result.Title)
					return
				}

				if result.Lat != -20 {
					t.Errorf("expected lat -20, got '%f'", result.Lat)
					return
				}

				if result.Long != 100 {
					t.Errorf("expected long 100, got '%f'", result.Long)
					return
				}

				if result.Radius != 55 {
					t.Errorf("expected radius 55, got '%d'", result.Radius)
					return
				}
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				infoLogs := logger.GetInfoLogs()
				if len(infoLogs) != 2 {
					t.Error("expected 2 info logs, got none")
					return
				}
			},
		},
		{
			name: "repository error",
			input: &CreateIncidentRequestInput{
				Title:  "Color",
				Lat:    -20,
				Long:   100,
				Radius: 55,
			},
			incidents: func() *mockIncidentsRepository {
				return &mockIncidentsRepository{
					createFunc: func(ctx context.Context, incedent *domain.Incident) error {
						return errors.New("failed database connection")
					},
				}
			},
			wantErr: true,
			validateLogs: func(t *testing.T, logger *mockLogger) {
				errorsLogs := logger.GetErrorLogs()
				if len(errorsLogs) != 1 {
					t.Error("expected 1 error log, got none")
					return
				}

				expectedMsg := "create incident request repository error"
				if errorsLogs[0].msg != expectedMsg {
					t.Errorf("expected error message '%s' not found", expectedMsg)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockLog := &mockLogger{}

			service := &Service{
				incidents: tt.incidents(),
				logger:    mockLog,
			}

			result, err := service.CreateIncident(ctx, tt.input)

			if tt.validateLogs != nil {
				tt.validateLogs(t, mockLog)
			}

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Fatal("expected result, got nil")
			}

			if tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

func TestService_GetIncidentByID(t *testing.T) {
	tests := []struct {
		name           string
		rawID          string
		incidents      func() *mockIncidentsRepository
		wantErr        bool
		cache          func() *mockCache
		validateResult func(t *testing.T, result *domain.Incident)
		validateLogs   func(t *testing.T, logger *mockLogger)
	}{
		{
			name:  "validation error - id is invalid",
			rawID: "id",
			incidents: func() *mockIncidentsRepository {
				return &mockIncidentsRepository{}
			},
			cache: func() *mockCache {
				return &mockCache{}
			},
			wantErr: true,
			validateLogs: func(t *testing.T, logger *mockLogger) {
				errorsLogs := logger.GetErrorLogs()
				if len(errorsLogs) != 1 {
					t.Error("expected 1 error log, got none")
					return
				}

				expectedMsg := "get incident by id validation failed"
				if errorsLogs[0].msg != expectedMsg {
					t.Errorf("expected error message '%s' not found", expectedMsg)
				}
			},
		},
		{
			name:  "success - get from cache",
			rawID: "1",
			incidents: func() *mockIncidentsRepository {
				return &mockIncidentsRepository{}
			},
			cache: func() *mockCache {
				return &mockCache{
					getFunc: func(ctx context.Context, key string) ([]byte, error) {
						incident := domain.Incident{
							ID:    1,
							Title: "Color",
						}

						dataByte, _ := json.Marshal(incident)
						return dataByte, nil
					},
				}
			},
			validateResult: func(t *testing.T, result *domain.Incident) {
				if result.ID != 1 {
					t.Errorf("expected id = 1, got '%d'", result.ID)
					return
				}

				if result.Title != "Color" {
					t.Errorf("expected 'Color', got '%s'", result.Description)
					return
				}
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				infoLogs := logger.GetInfoLogs()
				if len(infoLogs) != 2 {
					t.Error("expected 2 info log")
					return
				}

				expectedMsg := "successfully got from cache"
				if infoLogs[1].msg != expectedMsg {
					t.Errorf("expected msg '%s' not found", expectedMsg)
				}
			},
		},
		{
			name:  "success - get from db",
			rawID: "1",
			incidents: func() *mockIncidentsRepository {
				return &mockIncidentsRepository{
					getByIDFunc: func(ctx context.Context, id int) (*domain.Incident, error) {
						incident := domain.Incident{
							ID:    1,
							Title: "Color",
						}
						return &incident, nil
					},
				}
			},
			cache: func() *mockCache {
				return &mockCache{
					getFunc: func(ctx context.Context, key string) ([]byte, error) {
						return nil, nil
					},
				}
			},
			validateResult: func(t *testing.T, result *domain.Incident) {
				if result.ID != 1 {
					t.Errorf("expected id = 1, got '%d'", result.ID)
					return
				}

				if result.Title != "Color" {
					t.Errorf("expected 'Color', got '%s'", result.Description)
					return
				}
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				infoLogs := logger.GetInfoLogs()
				if len(infoLogs) != 3 {
					t.Error("expected 3 info log")
					return
				}

				expectedMsg := "successfully saved to cache"
				if infoLogs[1].msg != expectedMsg {
					t.Errorf("expected msg '%s' not found", expectedMsg)
				}
			},
		},
		{
			name:  "success - failed to save to cache",
			rawID: "1",
			incidents: func() *mockIncidentsRepository {
				return &mockIncidentsRepository{
					getByIDFunc: func(ctx context.Context, id int) (*domain.Incident, error) {
						incident := domain.Incident{
							ID:    1,
							Title: "Color",
						}
						return &incident, nil
					},
				}
			},
			cache: func() *mockCache {
				return &mockCache{
					getFunc: func(ctx context.Context, key string) ([]byte, error) {
						return nil, nil
					},
					saveFunc: func(ctx context.Context, data []byte, key string) error {
						return errors.New("failed cache connection")
					},
				}
			},
			validateResult: func(t *testing.T, result *domain.Incident) {
				if result.ID != 1 {
					t.Errorf("expected id = 1, got '%d'", result.ID)
					return
				}

				if result.Title != "Color" {
					t.Errorf("expected 'Color', got '%s'", result.Description)
					return
				}
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				errorsLogs := logger.GetErrorLogs()
				if len(errorsLogs) != 1 {
					t.Error("expected 1 error log")
					return
				}

				expectedMsg := "failed to save incident to cache"
				if errorsLogs[0].msg != expectedMsg {
					t.Errorf("expected msg '%s' not found", expectedMsg)
					return
				}

				infoLogs := logger.GetInfoLogs()
				if len(infoLogs) != 2 {
					t.Error("expected 2 info log")
					return
				}
			},
		},
		{
			name:  "repository error",
			rawID: "1",
			incidents: func() *mockIncidentsRepository {
				return &mockIncidentsRepository{
					getByIDFunc: func(ctx context.Context, id int) (*domain.Incident, error) {
						return nil, errors.New("failed database connection")
					},
				}
			},
			cache: func() *mockCache {
				return &mockCache{
					getFunc: func(ctx context.Context, key string) ([]byte, error) {
						return nil, nil
					},
				}
			},
			wantErr: true,
			validateLogs: func(t *testing.T, logger *mockLogger) {
				errorsLogs := logger.GetErrorLogs()
				if len(errorsLogs) != 1 {
					t.Error("expected 1 error log")
					return
				}

				expectedMsg := "get by id incident repository error"
				if errorsLogs[0].msg != expectedMsg {
					t.Errorf("expected msg '%s' not found", expectedMsg)
					return
				}

				infoLogs := logger.GetInfoLogs()
				if len(infoLogs) != 1 {
					t.Error("expected 1 info log")
					return
				}
			},
		},
		{
			name:  "success - failed to get from cache",
			rawID: "1",
			incidents: func() *mockIncidentsRepository {
				return &mockIncidentsRepository{
					getByIDFunc: func(ctx context.Context, id int) (*domain.Incident, error) {
						incident := domain.Incident{
							ID:    1,
							Title: "Color",
						}
						return &incident, nil
					},
				}
			},
			cache: func() *mockCache {
				return &mockCache{
					getFunc: func(ctx context.Context, key string) ([]byte, error) {
						return nil, errors.New("error")
					},
					saveFunc: func(ctx context.Context, data []byte, key string) error {
						return nil
					},
				}
			},
			validateResult: func(t *testing.T, result *domain.Incident) {
				if result.ID != 1 {
					t.Errorf("expected id = 1, got '%d'", result.ID)
					return
				}

				if result.Title != "Color" {
					t.Errorf("expected 'Color', got '%s'", result.Description)
					return
				}
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				warnsLogs := logger.GetWarnLogs()
				if len(warnsLogs) != 1 {
					t.Error("expected 1 warn log")
					return
				}

				expectedMsg := "failed to get incident from cache"
				if warnsLogs[0].msg != expectedMsg {
					t.Errorf("expected msg '%s' not found", expectedMsg)
					return
				}

				infoLogs := logger.GetInfoLogs()
				if len(infoLogs) != 3 {
					t.Error("expected 3 info log")
					return
				}
			},
		},
		{
			name:  "success - failed to get from cache (failed unmarshal)",
			rawID: "1",
			incidents: func() *mockIncidentsRepository {
				return &mockIncidentsRepository{
					getByIDFunc: func(ctx context.Context, id int) (*domain.Incident, error) {
						incident := domain.Incident{
							ID:    1,
							Title: "Color",
						}
						return &incident, nil
					},
				}
			},
			cache: func() *mockCache {
				return &mockCache{
					getFunc: func(ctx context.Context, key string) ([]byte, error) {
						return []byte("something"), nil
					},
					saveFunc: func(ctx context.Context, data []byte, key string) error {
						return nil
					},
				}
			},
			validateResult: func(t *testing.T, result *domain.Incident) {
				if result.ID != 1 {
					t.Errorf("expected id = 1, got '%d'", result.ID)
					return
				}

				if result.Title != "Color" {
					t.Errorf("expected 'Color', got '%s'", result.Description)
					return
				}
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				errorsLogs := logger.GetErrorLogs()
				if len(errorsLogs) != 1 {
					t.Error("expected 1 error log")
					return
				}

				expectedMsg := "failed to unmarshal incident from cache"
				if errorsLogs[0].msg != expectedMsg {
					t.Errorf("expected msg '%s' not found", expectedMsg)
					return
				}

				infoLogs := logger.GetInfoLogs()
				if len(infoLogs) != 3 {
					t.Error("expected 3 info log")
					return
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockLog := &mockLogger{}

			service := &Service{
				incidents: tt.incidents(),
				cache:     tt.cache(),
				logger:    mockLog,
			}

			result, err := service.GetIncidentByID(ctx, tt.rawID)

			if tt.validateLogs != nil {
				tt.validateLogs(t, mockLog)
			}

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Fatal("expected result, got nil")
			}

			if tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

func TestService_PaginateIncident(t *testing.T) {
	tests := []struct {
		name           string
		rawLimit       string
		rawPage        string
		incidents      func() *mockIncidentsRepository
		wantErr        bool
		validateResult func(t *testing.T, result *PaginateIncidentsOutput)
		validateLogs   func(t *testing.T, logger *mockLogger)
	}{
		{
			name:     "validation error - limit is invalid",
			rawLimit: "1-2",
			incidents: func() *mockIncidentsRepository {
				return &mockIncidentsRepository{}
			},
			wantErr: true,
			validateLogs: func(t *testing.T, logger *mockLogger) {
				errorsLogs := logger.GetErrorLogs()
				if len(errorsLogs) != 1 {
					t.Error("expected 1 error log")
					return
				}

				expectedMsg := "paginate incidents validation failed"
				if errorsLogs[0].msg != expectedMsg {
					t.Errorf("expected msg '%s' not found", expectedMsg)
				}
			},
		},
		{
			name:     "validation error - page is invalid",
			rawLimit: "1",
			rawPage:  "1-2",
			incidents: func() *mockIncidentsRepository {
				return &mockIncidentsRepository{}
			},
			wantErr: true,
			validateLogs: func(t *testing.T, logger *mockLogger) {
				errorsLogs := logger.GetErrorLogs()
				if len(errorsLogs) != 1 {
					t.Error("expected 1 error log")
					return
				}

				expectedMsg := "paginate incidents validation failed"
				if errorsLogs[0].msg != expectedMsg {
					t.Errorf("expected msg '%s' not found", expectedMsg)
				}
			},
		},
		{
			name:     "repository error",
			rawLimit: "5",
			rawPage:  "1",
			incidents: func() *mockIncidentsRepository {
				return &mockIncidentsRepository{
					paginateFunc: func(ctx context.Context, limit, offset int) ([]domain.Incident, int, error) {
						return nil, 0, errors.New("failed database connection")
					},
				}
			},
			wantErr: true,
			validateLogs: func(t *testing.T, logger *mockLogger) {
				infoLogs := logger.GetInfoLogs()
				if len(infoLogs) != 1 {
					t.Error("expected 1 info log")
					return
				}

				errorsLogs := logger.GetErrorLogs()
				if len(errorsLogs) != 1 {
					t.Error("expected 1 error log")
					return
				}

				expectedMsg := "paginate repository error"
				if errorsLogs[0].msg != expectedMsg {
					t.Errorf("expected msg '%s' not found", expectedMsg)
				}
			},
		},
		{
			name:     "success",
			rawLimit: "5",
			rawPage:  "1",
			incidents: func() *mockIncidentsRepository {
				return &mockIncidentsRepository{
					paginateFunc: func(ctx context.Context, limit, offset int) ([]domain.Incident, int, error) {
						incidents := []domain.Incident{
							{ID: 1, Title: "Color1"},
							{ID: 2, Title: "Color2"},
						}
						return incidents, 2, nil
					},
				}
			},
			validateResult: func(t *testing.T, result *PaginateIncidentsOutput) {
				if result.Incidents[0].ID != 1 || result.Incidents[0].Title != "Color1" {
					t.Error("expected other result")
					return
				}

				if result.Incidents[1].ID != 2 || result.Incidents[1].Title != "Color2" {
					t.Error("expected other result")
					return
				}

				if result.Pagination.Total != 2 {
					t.Error("expected total = 2")
					return
				}

				if result.Pagination.Limit != 5 || result.Pagination.Page != 1 {
					t.Error("expected limit = 5 and page = 1")
					return
				}
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				infoLogs := logger.GetInfoLogs()
				if len(infoLogs) != 2 {
					t.Error("expected 2 info log")
					return
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockLog := &mockLogger{}

			service := &Service{
				incidents: tt.incidents(),
				logger:    mockLog,
			}

			result, err := service.PaginateIncident(ctx, tt.rawLimit, tt.rawPage)

			if tt.validateLogs != nil {
				tt.validateLogs(t, mockLog)
			}

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if result == nil {
				t.Fatal("expected result, got nil")
				return
			}

			if tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

func TestService_DeleteIncident(t *testing.T) {
	tests := []struct {
		name         string
		rawID        string
		incidents    func() *mockIncidentsRepository
		cache        func() *mockCache
		wantErr      bool
		validateLogs func(t *testing.T, logger *mockLogger)
	}{
		{
			name:  "validation error - id is invalid",
			rawID: "1-2",
			incidents: func() *mockIncidentsRepository {
				return &mockIncidentsRepository{}
			},
			cache: func() *mockCache {
				return &mockCache{}
			},
			wantErr: true,
			validateLogs: func(t *testing.T, logger *mockLogger) {
				errorsLogs := logger.GetErrorLogs()
				if len(errorsLogs) != 1 {
					t.Error("expected 1 error log")
					return
				}

				expectedMsg := "delete incident validation failed"
				if errorsLogs[0].msg != expectedMsg {
					t.Errorf("expected msg '%s' not found", expectedMsg)
				}
			},
		},
		{
			name:  "repository error",
			rawID: "1",
			incidents: func() *mockIncidentsRepository {
				return &mockIncidentsRepository{
					deleteFunc: func(ctx context.Context, id int) error {
						return errors.New("failed database connection")
					},
				}
			},
			cache: func() *mockCache {
				return &mockCache{
					deleteFunc: func(ctx context.Context, key string) (bool, error) {
						return true, nil
					},
				}
			},
			wantErr: true,
			validateLogs: func(t *testing.T, logger *mockLogger) {
				infoLogs := logger.GetInfoLogs()
				if len(infoLogs) != 2 {
					t.Error("expected 2 info log")
					return
				}

				errorsLogs := logger.GetErrorLogs()
				if len(errorsLogs) != 1 {
					t.Error("expected 1 error log")
					return
				}

				expectedMsg := "delete repository error"
				if errorsLogs[0].msg != expectedMsg {
					t.Errorf("expected msg '%s' not found", expectedMsg)
				}
			},
		},
		{
			name:  "success - but cache error",
			rawID: "1",
			incidents: func() *mockIncidentsRepository {
				return &mockIncidentsRepository{
					deleteFunc: func(ctx context.Context, id int) error {
						return nil
					},
				}
			},
			cache: func() *mockCache {
				return &mockCache{
					deleteFunc: func(ctx context.Context, key string) (bool, error) {
						return false, errors.New("error")
					},
				}
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				infoLogs := logger.GetInfoLogs()
				if len(infoLogs) != 2 {
					t.Error("expected 2 info log")
					return
				}

				errorsLogs := logger.GetErrorLogs()
				if len(errorsLogs) != 1 {
					t.Error("expected 1 error log")
					return
				}

				expectedMsg := "failed to delete incident from cache"
				if errorsLogs[0].msg != expectedMsg {
					t.Errorf("expected msg '%s' not found", expectedMsg)
				}
			},
		},
		{
			name:  "success",
			rawID: "1",
			incidents: func() *mockIncidentsRepository {
				return &mockIncidentsRepository{
					deleteFunc: func(ctx context.Context, id int) error {
						return nil
					},
				}
			},
			cache: func() *mockCache {
				return &mockCache{
					deleteFunc: func(ctx context.Context, key string) (bool, error) {
						return true, nil
					},
				}
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				infoLogs := logger.GetInfoLogs()
				if len(infoLogs) != 3 {
					t.Error("expected 3 info log")
					return
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockLog := &mockLogger{}

			service := &Service{
				incidents: tt.incidents(),
				cache:     tt.cache(),
				logger:    mockLog,
			}

			err := service.DeleteIncident(ctx, tt.rawID)

			if tt.validateLogs != nil {
				tt.validateLogs(t, mockLog)
			}

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Error("unexpected error")
				return
			}
		})
	}
}

func TestService_FullUpdateIncident(t *testing.T) {
	tests := []struct {
		name           string
		input          *FullUpdateIncidentRequestInput
		incidents      func() *mockIncidentsRepository
		cache          func() *mockCache
		wantErr        bool
		validateResult func(t *testing.T, result *domain.Incident)
		validateLogs   func(t *testing.T, logger *mockLogger)
	}{
		{
			name: "validation error - id is invalid",
			input: &FullUpdateIncidentRequestInput{
				ID: "1-2",
			},
			incidents: func() *mockIncidentsRepository {
				return &mockIncidentsRepository{}
			},
			cache: func() *mockCache {
				return &mockCache{}
			},
			wantErr: true,
			validateLogs: func(t *testing.T, logger *mockLogger) {
				errorsLogs := logger.GetErrorLogs()
				if len(errorsLogs) != 1 {
					t.Error("expected 1 error log")
					return
				}

				expectedMsg := "full update incident request validation failed"
				if errorsLogs[0].msg != expectedMsg {
					t.Errorf("expected msg '%s' not found", expectedMsg)
				}
			},
		},
		{
			name: "validation error - title is invalid",
			input: &FullUpdateIncidentRequestInput{
				ID: "1",
			},
			incidents: func() *mockIncidentsRepository {
				return &mockIncidentsRepository{}
			},
			cache: func() *mockCache {
				return &mockCache{}
			},
			wantErr: true,
			validateLogs: func(t *testing.T, logger *mockLogger) {
				errorsLogs := logger.GetErrorLogs()
				if len(errorsLogs) != 1 {
					t.Error("expected 1 error log")
					return
				}

				expectedMsg := "full update incident request validation failed"
				if errorsLogs[0].msg != expectedMsg {
					t.Errorf("expected msg '%s' not found", expectedMsg)
				}
			},
		},
		{
			name: "repostitory error",
			input: &FullUpdateIncidentRequestInput{
				ID:     "1",
				Title:  "Color",
				Lat:    50,
				Long:   50,
				Radius: 50,
			},
			incidents: func() *mockIncidentsRepository {
				return &mockIncidentsRepository{
					fullUpdateFunc: func(ctx context.Context, incident *domain.Incident) error {
						return errors.New("error")
					},
				}
			},
			cache: func() *mockCache {
				return &mockCache{}
			},
			wantErr: true,
			validateLogs: func(t *testing.T, logger *mockLogger) {
				infoLogs := logger.GetInfoLogs()
				if len(infoLogs) != 1 {
					t.Error("expected 1 info log")
					return
				}

				errorsLogs := logger.GetErrorLogs()
				if len(errorsLogs) != 1 {
					t.Error("expected 1 error log")
					return
				}

				expectedMsg := "full update incident request repository error"
				if errorsLogs[0].msg != expectedMsg {
					t.Errorf("expected msg '%s' not found", expectedMsg)
				}
			},
		},
		{
			name: "success - but cache error",
			input: &FullUpdateIncidentRequestInput{
				ID:     "1",
				Title:  "Color",
				Lat:    50,
				Long:   50,
				Radius: 50,
			},
			incidents: func() *mockIncidentsRepository {
				return &mockIncidentsRepository{
					fullUpdateFunc: func(ctx context.Context, incident *domain.Incident) error {
						return nil
					},
				}
			},
			cache: func() *mockCache {
				return &mockCache{
					deleteFunc: func(ctx context.Context, key string) (bool, error) {
						return false, errors.New("error")
					},
				}
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				infoLogs := logger.GetInfoLogs()
				if len(infoLogs) != 2 {
					t.Error("expected 2 info log")
					return
				}

				errorsLogs := logger.GetErrorLogs()
				if len(errorsLogs) != 1 {
					t.Error("expected 1 error log")
					return
				}

				expectedMsg := "failed to delete incident from cache"
				if errorsLogs[0].msg != expectedMsg {
					t.Errorf("expected msg '%s' not found", expectedMsg)
				}
			},
		},
		{
			name: "success",
			input: &FullUpdateIncidentRequestInput{
				ID:     "1",
				Title:  "Color",
				Lat:    50,
				Long:   50,
				Radius: 50,
			},
			incidents: func() *mockIncidentsRepository {
				return &mockIncidentsRepository{
					fullUpdateFunc: func(ctx context.Context, incident *domain.Incident) error {
						return nil
					},
				}
			},
			cache: func() *mockCache {
				return &mockCache{
					deleteFunc: func(ctx context.Context, key string) (bool, error) {
						return true, nil
					},
				}
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				infoLogs := logger.GetInfoLogs()
				if len(infoLogs) != 3 {
					t.Error("expected 3 info log")
					return
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			mockLog := &mockLogger{}

			service := &Service{
				incidents: tt.incidents(),
				cache:     tt.cache(),
				logger:    mockLog,
			}

			result, err := service.FullUpdateIncident(ctx, tt.input)

			if tt.validateLogs != nil {
				tt.validateLogs(t, mockLog)
			}

			if tt.wantErr {
				if err == nil {
					t.Error("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Error("unexpected error")
				return
			}

			if result == nil {
				t.Fatal("expected result, got nil")
				return
			}

			if tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

func ptr[T any](v T) *T { return &v }
