package service

import (
	"context"
	"encoding/json"
	"errors"
	"red_collar/internal/domain"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestService_CreateIncident(t *testing.T) {
	tests := []struct {
		name           string
		input          *CreateIncidentRequestInput
		incidents      func() *mockIncidentsRepository
		wantErr        bool
		errType        func(err error) bool
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
			errType: func(err error) bool {
				var appErr *domain.AppError
				return errors.As(err, &appErr) && appErr.Code == domain.CodeInvalidValidation
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				errorLogs := logger.GetErrorLogs()
				require.Len(t, errorLogs, 1, "should have exactly 1 error log")
				require.Equal(t, "create incident request validation failed", errorLogs[0].msg)
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
			errType: func(err error) bool {
				var appErr *domain.AppError
				return errors.As(err, &appErr) && appErr.Code == domain.CodeInvalidValidation
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				errorLogs := logger.GetErrorLogs()
				require.Len(t, errorLogs, 1)
				require.Equal(t, "create incident request validation failed", errorLogs[0].msg)
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
			errType: func(err error) bool {
				var appErr *domain.AppError
				return errors.As(err, &appErr) && appErr.Code == domain.CodeInvalidValidation
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				errorLogs := logger.GetErrorLogs()
				require.Len(t, errorLogs, 1)
				require.Equal(t, "create incident request validation failed", errorLogs[0].msg)
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
				require.NotNil(t, result)
				require.True(t, result.Active)
				require.Equal(t, "without description", result.Description)
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				infoLogs := logger.GetInfoLogs()
				require.Len(t, infoLogs, 2, "should log attempt and success")
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
				require.NotNil(t, result)
				require.True(t, result.Active)
				require.Equal(t, "something", result.Description)
				require.Equal(t, "Color", result.Title)
				require.Equal(t, -20.0, result.Lat)
				require.Equal(t, 100.0, result.Long)
				require.Equal(t, 55, result.Radius)
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				infoLogs := logger.GetInfoLogs()
				require.Len(t, infoLogs, 2, "should log attempt and success")
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
				errorLogs := logger.GetErrorLogs()
				require.Len(t, errorLogs, 1, "should log repository error")
				require.Equal(t, "create incident request repository error", errorLogs[0].msg)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

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
				require.Error(t, err, "expected error but got none")
				if tt.errType != nil {
					require.True(t, tt.errType(err), "wrong error type: %v", err)
				}
				require.Nil(t, result, "result should be nil on error")
				return
			}

			require.NoError(t, err, "unexpected error: %v", err)
			require.NotNil(t, result, "expected result, got nil")

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
		errType        func(err error) bool
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
			errType: func(err error) bool {
				var appErr *domain.AppError
				return errors.As(err, &appErr) && appErr.Code == domain.CodeInvalidValidation
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				errorLogs := logger.GetErrorLogs()
				require.Len(t, errorLogs, 1, "should have exactly 1 error log")
				require.Equal(t, "get incident by id validation failed", errorLogs[0].msg)
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
				require.NotNil(t, result)
				require.Equal(t, 1, result.ID)
				require.Equal(t, "Color", result.Title)
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				infoLogs := logger.GetInfoLogs()
				require.Len(t, infoLogs, 2, "should log attempt and cache hit")
				require.Equal(t, "successfully got from cache", infoLogs[1].msg)
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
				require.NotNil(t, result)
				require.Equal(t, 1, result.ID)
				require.Equal(t, "Color", result.Title)
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				infoLogs := logger.GetInfoLogs()
				require.Len(t, infoLogs, 3, "should log attempt, save to cache, and success")
				require.Equal(t, "successfully saved to cache", infoLogs[1].msg)
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
				require.NotNil(t, result)
				require.Equal(t, 1, result.ID)
				require.Equal(t, "Color", result.Title)
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				errorLogs := logger.GetErrorLogs()
				require.Len(t, errorLogs, 1, "should log cache save error")
				require.Equal(t, "failed to save incident to cache", errorLogs[0].msg)

				infoLogs := logger.GetInfoLogs()
				require.Len(t, infoLogs, 2, "should log attempt and success")
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
				errorLogs := logger.GetErrorLogs()
				require.Len(t, errorLogs, 1, "should log repository error")
				require.Equal(t, "get by id incident repository error", errorLogs[0].msg)

				infoLogs := logger.GetInfoLogs()
				require.Len(t, infoLogs, 1, "should log attempt")
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
				require.NotNil(t, result)
				require.Equal(t, 1, result.ID)
				require.Equal(t, "Color", result.Title)
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				warnLogs := logger.GetWarnLogs()
				require.Len(t, warnLogs, 1, "should log cache get warning")
				require.Equal(t, "failed to get incident from cache", warnLogs[0].msg)

				infoLogs := logger.GetInfoLogs()
				require.Len(t, infoLogs, 3, "should log attempt, save to cache, and success")
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
				require.NotNil(t, result)
				require.Equal(t, 1, result.ID)
				require.Equal(t, "Color", result.Title)
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				errorLogs := logger.GetErrorLogs()
				require.Len(t, errorLogs, 1, "should log unmarshal error")
				require.Equal(t, "failed to unmarshal incident from cache", errorLogs[0].msg)

				infoLogs := logger.GetInfoLogs()
				require.Len(t, infoLogs, 3, "should log attempt, save to cache, and success")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

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
				require.Error(t, err, "expected error but got none")
				if tt.errType != nil {
					require.True(t, tt.errType(err), "wrong error type: %v", err)
				}
				require.Nil(t, result, "result should be nil on error")
				return
			}

			require.NoError(t, err, "unexpected error: %v", err)
			require.NotNil(t, result, "expected result, got nil")

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
		errType        func(err error) bool
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
			errType: func(err error) bool {
				var appErr *domain.AppError
				return errors.As(err, &appErr) && appErr.Code == domain.CodeInvalidValidation
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				errorLogs := logger.GetErrorLogs()
				require.Len(t, errorLogs, 1, "should have exactly 1 error log")
				require.Equal(t, "paginate incidents validation failed", errorLogs[0].msg)
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
			errType: func(err error) bool {
				var appErr *domain.AppError
				return errors.As(err, &appErr) && appErr.Code == domain.CodeInvalidValidation
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				errorLogs := logger.GetErrorLogs()
				require.Len(t, errorLogs, 1)
				require.Equal(t, "paginate incidents validation failed", errorLogs[0].msg)
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
				require.Len(t, infoLogs, 1, "should log attempt")

				errorLogs := logger.GetErrorLogs()
				require.Len(t, errorLogs, 1, "should log repository error")
				require.Equal(t, "paginate repository error", errorLogs[0].msg)
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
				require.NotNil(t, result)
				require.NotNil(t, result.Incidents)
				require.Len(t, result.Incidents, 2)

				require.Equal(t, 1, result.Incidents[0].ID)
				require.Equal(t, "Color1", result.Incidents[0].Title)

				require.Equal(t, 2, result.Incidents[1].ID)
				require.Equal(t, "Color2", result.Incidents[1].Title)

				require.NotNil(t, result.Pagination)
				require.Equal(t, 2, result.Pagination.Total)
				require.Equal(t, 5, result.Pagination.Limit)
				require.Equal(t, 1, result.Pagination.Page)
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				infoLogs := logger.GetInfoLogs()
				require.Len(t, infoLogs, 2, "should log attempt and success")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

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
				require.Error(t, err, "expected error but got none")
				if tt.errType != nil {
					require.True(t, tt.errType(err), "wrong error type: %v", err)
				}
				require.Nil(t, result, "result should be nil on error")
				return
			}

			require.NoError(t, err, "unexpected error: %v", err)
			require.NotNil(t, result, "expected result, got nil")

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
		errType      func(err error) bool
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
			errType: func(err error) bool {
				var appErr *domain.AppError
				return errors.As(err, &appErr) && appErr.Code == domain.CodeInvalidValidation
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				errorLogs := logger.GetErrorLogs()
				require.Len(t, errorLogs, 1, "should have exactly 1 error log")
				require.Equal(t, "delete incident validation failed", errorLogs[0].msg)
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
				require.Len(t, infoLogs, 2, "should log attempt")

				errorLogs := logger.GetErrorLogs()
				require.Len(t, errorLogs, 1, "should log repository error")
				require.Equal(t, "delete repository error", errorLogs[0].msg)
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
				require.Len(t, infoLogs, 2, "should log attempt and success")

				errorLogs := logger.GetErrorLogs()
				require.Len(t, errorLogs, 1, "should log cache delete error")
				require.Equal(t, "failed to delete incident from cache", errorLogs[0].msg)
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
				require.Len(t, infoLogs, 3, "should log attempt, cache delete, and success")
				errorLogs := logger.GetErrorLogs()
				require.Empty(t, errorLogs, "should not have errors")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

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
				require.Error(t, err, "expected error but got none")
				if tt.errType != nil {
					require.True(t, tt.errType(err), "wrong error type: %v", err)
				}
				return
			}

			require.NoError(t, err, "unexpected error: %v", err)
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
		errType        func(err error) bool
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
			errType: func(err error) bool {
				var appErr *domain.AppError
				return errors.As(err, &appErr) && appErr.Code == domain.CodeInvalidValidation
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				errorLogs := logger.GetErrorLogs()
				require.Len(t, errorLogs, 1, "should have exactly 1 error log")
				require.Equal(t, "full update incident request validation failed", errorLogs[0].msg)
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
			errType: func(err error) bool {
				var appErr *domain.AppError
				return errors.As(err, &appErr) && appErr.Code == domain.CodeInvalidValidation
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				errorLogs := logger.GetErrorLogs()
				require.Len(t, errorLogs, 1)
				require.Equal(t, "full update incident request validation failed", errorLogs[0].msg)
			},
		},
		{
			name: "repository error",
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
				require.Len(t, infoLogs, 1, "should log attempt")

				errorLogs := logger.GetErrorLogs()
				require.Len(t, errorLogs, 1, "should log repository error")
				require.Equal(t, "full update incident request repository error", errorLogs[0].msg)
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
				require.Len(t, infoLogs, 2, "should log attempt and success")

				errorLogs := logger.GetErrorLogs()
				require.Len(t, errorLogs, 1, "should log cache delete error")
				require.Equal(t, "failed to delete incident from cache", errorLogs[0].msg)
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
				require.Len(t, infoLogs, 3, "should log attempt, cache delete, and success")
				errorLogs := logger.GetErrorLogs()
				require.Empty(t, errorLogs, "should not have errors")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

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
				require.Error(t, err, "expected error but got none")
				if tt.errType != nil {
					require.True(t, tt.errType(err), "wrong error type: %v", err)
				}
				require.Nil(t, result, "result should be nil on error")
				return
			}

			require.NoError(t, err, "unexpected error: %v", err)
			require.NotNil(t, result, "expected result, got nil")

			if tt.validateResult != nil {
				tt.validateResult(t, result)
			}
		})
	}
}

func ptr[T any](v T) *T { return &v }
