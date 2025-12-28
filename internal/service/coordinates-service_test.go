package service

import (
	"context"
	"errors"
	"red_collar/internal/domain"
	"testing"
	"time"
)

func TestService_CheckCoordinates(t *testing.T) {
	tests := []struct {
		name            string
		input           *CheckCoordinatesRequestInput
		coordinatesMock func() *mockCoordinatesRepository
		queueMock       func() *mockQueue
		wantErr         bool
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
			validateLogs: func(t *testing.T, logger *mockLogger) {
				errorsLogs := logger.GetErrorLogs()
				if len(errorsLogs) != 1 {
					t.Error("expected 1 error log, got none")
					return
				}

				expectedMsg := "check coordinates request validation failed"
				if errorsLogs[0].msg != expectedMsg {
					t.Errorf("expected error message '%s' not found", expectedMsg)
				}
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
			validateLogs: func(t *testing.T, logger *mockLogger) {
				errorsLogs := logger.GetErrorLogs()
				if len(errorsLogs) != 1 {
					t.Error("expected 1 error log, got none")
					return
				}

				expectedMsg := "check coordinates request validation failed"
				if errorsLogs[0].msg != expectedMsg {
					t.Errorf("expected error message '%s' not found", expectedMsg)
				}
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
				if len(infoLogs) != 1 {
					t.Error("expected 1 info log, got none")
					return
				}

				errorsLogs := logger.GetErrorLogs()
				if len(errorsLogs) != 1 {
					t.Error("expected 1 error log, got none")
					return
				}

				expectedMsg := "check coordinates request repository error"
				if errorsLogs[0].msg != expectedMsg {
					t.Errorf("expected error message '%s' not found", expectedMsg)
				}
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
				if result.UserID != "colorvax" {
					t.Errorf("expected UserID 'colorvax', got '%s'", result.UserID)
				}
				if result.Lat != 45 {
					t.Errorf("expected Lat 45, got %f", result.Lat)
				}
				if result.Long != 30 {
					t.Errorf("expected Long 30, got %f", result.Long)
				}
				if result.InDangerZone != false {
					t.Errorf("expected inDangerZone false, got %v", result.InDangerZone)
				}
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				infoLogs := logger.GetInfoLogs()
				if len(infoLogs) != 2 {
					t.Error("expected 2 info log, got none")
					return
				}
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
			wantErr: false,
			validateResult: func(t *testing.T, result *domain.LocationCheck) {
				if result.UserID != "colorvax" {
					t.Errorf("expected UserID 'colorvax', got '%s'", result.UserID)
				}
				if result.InDangerZone != true {
					t.Errorf("expected inDangerZone true, got %v", result.InDangerZone)
				}
				if result.NearestID == nil || *result.NearestID != 10 {
					t.Errorf("expected NearestID 10, got %v", result.NearestID)
				}
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				infoLogs := logger.GetInfoLogs()
				if len(infoLogs) != 3 {
					t.Error("expected 3 info log, got none")
					return
				}
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
			wantErr: false,
			validateResult: func(t *testing.T, result *domain.LocationCheck) {
				if result.UserID != "colorvax" {
					t.Errorf("expected UserID 'colorvax', got '%s'", result.UserID)
				}
				if result.InDangerZone != true {
					t.Errorf("expected inDangerZone true, got %v", result.InDangerZone)
				}
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				infoLogs := logger.GetInfoLogs()
				if len(infoLogs) != 2 {
					t.Error("expected 2 info log, got none")
					return
				}

				errorsLogs := logger.GetErrorLogs()
				if len(errorsLogs) != 1 {
					t.Error("expectedd 1 error log, got none")
					return
				}

				expectedMsg := "failed to enqueue webhook task"
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
				coordinates: tt.coordinatesMock(),
				queue:       tt.queueMock(),
				logger:      mockLog,
			}

			result, err := service.CheckCoordinates(ctx, tt.input)

			if tt.validateLogs != nil {
				tt.validateLogs(t, mockLog)
			}

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
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

func TestService_GetStats(t *testing.T) {
	tests := []struct {
		name              string
		timeWindowMinutes int
		coordinates       func() *mockCoordinatesRepository
		wantErr           bool
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
			wantErr: false,
			validateResult: func(t *testing.T, result []domain.ZoneStat) {
				if len(result) != 2 {
					t.Errorf("expected len zones 2, got %d", len(result))
				}
				if result[0].ZoneID != 1 || result[0].UserCount != 5 {
					t.Errorf("expected ZoneID 1 and UserCount 5, got ZoneID %d and UserCount %d", result[0].ZoneID, result[0].UserCount)
				}
				if result[1].ZoneID != 2 || result[1].UserCount != 10 {
					t.Errorf("expected ZoneID 2 and UserCount 10, got ZoneID %d and UserCount %d", result[1].ZoneID, result[1].UserCount)
				}
			},
			validateLogs: func(t *testing.T, logger *mockLogger) {
				infoLogs := logger.GetInfoLogs()
				if len(infoLogs) != 2 {
					t.Error("expected 2 info log, got none")
					return
				}
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
				if len(infoLogs) != 1 {
					t.Error("expected 1 info log, got none")
					return
				}

				errorsLogs := logger.GetErrorLogs()
				if len(errorsLogs) != 1 {
					t.Error("expected 1 error log, got none")
					return
				}

				expectedMsg := "failed to get stat repository error"
				if errorsLogs[0].msg != expectedMsg {
					t.Errorf("expected error message '%s' not found", expectedMsg)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			service := &Service{
				coordinates: tt.coordinates(),
				logger:      &mockLogger{},
			}

			ctx := context.Background()

			zones, err := service.GetStats(ctx, tt.timeWindowMinutes)
			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if zones == nil {
				t.Fatalf("expected zones, got nil")
			}

			if tt.validateResult != nil {
				tt.validateResult(t, zones)
			}
		})
	}
}
