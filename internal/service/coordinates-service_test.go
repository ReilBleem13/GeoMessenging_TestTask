package service

import (
	"context"
	"errors"
	"red_collar/internal/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestService_CheckCoordinates(t *testing.T) {
	tests := []struct {
		name            string
		input           *CheckCoordinatesRequestInput
		coordinatesMock func() *mockCoordinatesRepository
		queueMock       func() *mockQueue
		wantErr         bool
		errType         func(err error) bool
		validateResult  func(t *testing.T, result *domain.LocationCheck)
		validateLogs    func(t *testing.T, logger *mockLogger)
	}{
		{
			name: "validation error - invalid latitude",
			input: &CheckCoordinatesRequestInput{
				UserID: "colorvax",
				Lat:    91,
				Long:   0,
			},
			coordinatesMock: func() *mockCoordinatesRepository {
				return &mockCoordinatesRepository{}
			},
			queueMock: func() *mockQueue {
				return &mockQueue{}
			},
			wantErr: true,
			errType: func(err error) bool {
				var appErr *domain.AppError
				if !errors.As(err, &appErr) {
					return false
				}
				return appErr.Code == domain.CodeInvalidValidation
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				errorLogs := logger.GetErrorLogs()
				require.Len(t, errorLogs, 1, "should have exactly 1 error log")
				require.Equal(t, "check coordinates request validation failed", errorLogs[0].msg)
			},
		},
		{
			name: "validation error - invalid longitude",
			input: &CheckCoordinatesRequestInput{
				UserID: "colorvax",
				Lat:    85,
				Long:   190,
			},
			coordinatesMock: func() *mockCoordinatesRepository {
				return &mockCoordinatesRepository{}
			},
			queueMock: func() *mockQueue {
				return &mockQueue{}
			},
			wantErr: true,
			errType: func(err error) bool {
				var appErr *domain.AppError
				if !errors.As(err, &appErr) {
					return false
				}
				return appErr.Code == domain.CodeInvalidValidation
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				errorLogs := logger.GetErrorLogs()
				require.Len(t, errorLogs, 1)
				require.Equal(t, "check coordinates request validation failed", errorLogs[0].msg)
			},
		},
		{
			name: "repository error",
			input: &CheckCoordinatesRequestInput{
				UserID: "colorvax",
				Lat:    85,
				Long:   160,
			},
			coordinatesMock: func() *mockCoordinatesRepository {
				return &mockCoordinatesRepository{
					checkFunc: func(ctx context.Context, locCheck *domain.LocationCheck) error {
						return errors.New("failed database connection")
					},
				}
			},
			queueMock: func() *mockQueue {
				return &mockQueue{}
			},
			wantErr: true,
			validateLogs: func(t *testing.T, logger *mockLogger) {
				infoLogs := logger.GetInfoLogs()
				require.Len(t, infoLogs, 1, "should log attempt")

				errorLogs := logger.GetErrorLogs()
				require.Len(t, errorLogs, 1, "should log repository error")
				require.Equal(t, "check coordinates request repository error", errorLogs[0].msg)
			},
		},
		{
			name: "success - not in danger zone",
			input: &CheckCoordinatesRequestInput{
				UserID: "colorvax",
				Lat:    45,
				Long:   30,
			},
			coordinatesMock: func() *mockCoordinatesRepository {
				return &mockCoordinatesRepository{
					checkFunc: func(ctx context.Context, locCheck *domain.LocationCheck) error {
						locCheck.ID = 1
						locCheck.CheckedAt = time.Now()
						locCheck.InDangerZone = false
						locCheck.NearestID = nil
						return nil
					},
				}
			},
			queueMock: func() *mockQueue {
				return &mockQueue{}
			},
			wantErr: false,
			validateResult: func(t *testing.T, result *domain.LocationCheck) {
				require.NotNil(t, result)
				require.Equal(t, "colorvax", result.UserID)
				require.Equal(t, 45.0, result.Lat)
				require.Equal(t, 30.0, result.Long)
				require.False(t, result.InDangerZone)
				require.Nil(t, result.NearestID)
				require.NotZero(t, result.ID)
				require.False(t, result.CheckedAt.IsZero())
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				infoLogs := logger.GetInfoLogs()
				require.Len(t, infoLogs, 2, "should log attempt and success")
			},
		},
		{
			name: "success - in danger zone, enqueue successful",
			input: &CheckCoordinatesRequestInput{
				UserID: "colorvax",
				Lat:    50,
				Long:   40,
			},
			coordinatesMock: func() *mockCoordinatesRepository {
				return &mockCoordinatesRepository{
					checkFunc: func(ctx context.Context, locCheck *domain.LocationCheck) error {
						locCheck.ID = 1
						locCheck.CheckedAt = time.Now()
						locCheck.InDangerZone = true
						nearestID := 10
						locCheck.NearestID = &nearestID
						return nil
					},
				}
			},
			queueMock: func() *mockQueue {
				return &mockQueue{
					enqueueFunc: func(ctx context.Context, check *domain.LocationCheck) error {

						return nil
					},
				}
			},
			validateResult: func(t *testing.T, result *domain.LocationCheck) {
				require.NotNil(t, result)
				require.Equal(t, "colorvax", result.UserID)
				require.True(t, result.InDangerZone)
				require.NotNil(t, result.NearestID)
				require.Equal(t, 10, *result.NearestID)
				require.Equal(t, 50.0, result.Lat)
				require.Equal(t, 40.0, result.Long)
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				infoLogs := logger.GetInfoLogs()
				require.Len(t, infoLogs, 3, "should log attempt, success, and enqueue")
				errorLogs := logger.GetErrorLogs()
				require.Empty(t, errorLogs, "should not have errors")
			},
		},
		{
			name: "success - in danger zone, enqueue failed",
			input: &CheckCoordinatesRequestInput{
				UserID: "colorvax",
				Lat:    50,
				Long:   40,
			},
			coordinatesMock: func() *mockCoordinatesRepository {
				return &mockCoordinatesRepository{
					checkFunc: func(ctx context.Context, locCheck *domain.LocationCheck) error {
						locCheck.ID = 1
						locCheck.CheckedAt = time.Now()
						locCheck.InDangerZone = true
						nearestID := 10
						locCheck.NearestID = &nearestID
						return nil
					},
				}
			},
			queueMock: func() *mockQueue {
				return &mockQueue{
					enqueueFunc: func(ctx context.Context, check *domain.LocationCheck) error {
						return errors.New("failed queue connection")
					},
				}
			},
			validateResult: func(t *testing.T, result *domain.LocationCheck) {
				require.NotNil(t, result)
				require.Equal(t, "colorvax", result.UserID)
				require.True(t, result.InDangerZone)
				require.NotNil(t, result.NearestID)
				require.Equal(t, 10, *result.NearestID)
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				infoLogs := logger.GetInfoLogs()
				require.Len(t, infoLogs, 2, "should log attempt and success (but not enqueue success)")

				errorLogs := logger.GetErrorLogs()
				require.Len(t, errorLogs, 1, "should log enqueue error")
				require.Equal(t, "failed to enqueue webhook task", errorLogs[0].msg)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			ctx := context.Background()
			mockLog := &mockLogger{}

			service := &Service{
				coordinates: tt.coordinatesMock(),
				queue:       tt.queueMock(),
				logger:      mockLog,
			}

			result, err := service.CheckCoordinates(ctx, tt.input)

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

func TestService_GetStats(t *testing.T) {
	tests := []struct {
		name              string
		timeWindowMinutes int
		coordinates       func() *mockCoordinatesRepository
		wantErr           bool
		errType           func(err error) bool
		validateResult    func(t *testing.T, result []domain.ZoneStat)
		validateLogs      func(t *testing.T, logger *mockLogger)
	}{
		{
			name:              "success",
			timeWindowMinutes: 10,
			coordinates: func() *mockCoordinatesRepository {
				return &mockCoordinatesRepository{
					getStatsFunc: func(ctx context.Context, timeWindowsMinutes int) ([]domain.ZoneStat, error) {
						zones := []domain.ZoneStat{
							{ZoneID: 1, UserCount: 5},
							{ZoneID: 2, UserCount: 10},
						}
						return zones, nil
					},
				}
			},
			validateResult: func(t *testing.T, result []domain.ZoneStat) {
				require.NotNil(t, result)
				require.Len(t, result, 2)

				require.Equal(t, 1, result[0].ZoneID)
				require.Equal(t, 5, result[0].UserCount)

				require.Equal(t, 2, result[1].ZoneID)
				require.Equal(t, 10, result[1].UserCount)
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				infoLogs := logger.GetInfoLogs()
				require.Len(t, infoLogs, 2, "should log attempt and success")
			},
		},
		{
			name:              "repository error",
			timeWindowMinutes: 10,
			coordinates: func() *mockCoordinatesRepository {
				return &mockCoordinatesRepository{
					getStatsFunc: func(ctx context.Context, timeWindowsMinutes int) ([]domain.ZoneStat, error) {
						return nil, errors.New("failed database connection")
					},
				}
			},
			wantErr: true,
			validateLogs: func(t *testing.T, logger *mockLogger) {
				infoLogs := logger.GetInfoLogs()
				require.Len(t, infoLogs, 1, "should log attempt")

				errorLogs := logger.GetErrorLogs()
				require.Len(t, errorLogs, 1, "should log repository error")
				require.Equal(t, "failed to get stat repository error", errorLogs[0].msg)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockLog := &mockLogger{}
			service := &Service{
				coordinates: tt.coordinates(),
				logger:      mockLog,
			}

			ctx := context.Background()

			zones, err := service.GetStats(ctx, tt.timeWindowMinutes)

			if tt.validateLogs != nil {
				tt.validateLogs(t, mockLog)
			}

			if tt.wantErr {
				require.Error(t, err, "expected error but got none")
				if tt.errType != nil {
					require.True(t, tt.errType(err), "wrong error type: %v", err)
				}
				require.Nil(t, zones, "zones should be nil on error")
				return
			}

			require.NoError(t, err, "unexpected error: %v", err)
			require.NotNil(t, zones, "expected zones, got nil")

			if tt.validateResult != nil {
				tt.validateResult(t, zones)
			}
		})
	}
}
