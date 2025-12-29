package repository

import (
	"context"
	"fmt"
	"red_collar/internal/domain"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestCoordinatesRepository_Check(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping intergration test")
	}

	if testDB == nil {
		setupTestDB(t)
	}

	ctx := context.Background()

	tests := []struct {
		name     string
		locCheck *domain.LocationCheck
		setup    func(t *testing.T)
		wantErr  bool
		validate func(t *testing.T, res *domain.LocationCheck)
	}{
		{
			name: "success",
			locCheck: &domain.LocationCheck{
				UserID: "colorvax",
				Lat:    50,
				Long:   50,
			},
			setup: func(t *testing.T) {
				incident := &domain.Incident{
					Title:       "Incident",
					Description: "Description",
					Lat:         50.0,
					Long:        50.1,
					Radius:      100,
					Active:      true,
				}
				err := testRepo.Create(ctx, incident)
				require.NoError(t, err)
			},
			validate: func(t *testing.T, res *domain.LocationCheck) {
				require.Equal(t, 1, res.ID)
				require.Nil(t, res.NearestID)
				require.Equal(t, "colorvax", res.UserID)
				require.Equal(t, 50.0, res.Lat)
				require.Equal(t, 50.0, res.Long)
				require.False(t, res.InDangerZone)
				require.WithinDuration(t, time.Now(), res.CheckedAt, 10*time.Second)
			},
		},
		{
			name: "success - in danger zone",
			locCheck: &domain.LocationCheck{
				UserID: "colorvax",
				Lat:    50,
				Long:   50,
			},
			setup: func(t *testing.T) {
				incident := &domain.Incident{
					Title:       "Incident",
					Description: "Description",
					Lat:         50.0,
					Long:        50.01,
					Radius:      1000,
					Active:      true,
				}
				err := testRepo.Create(ctx, incident)
				require.NoError(t, err)
			},
			validate: func(t *testing.T, res *domain.LocationCheck) {
				require.Equal(t, 1, res.ID)
				require.Equal(t, 1, *res.NearestID)
				require.Equal(t, "colorvax", res.UserID)
				require.Equal(t, 50.0, res.Lat)
				require.Equal(t, 50.0, res.Long)
				require.True(t, res.InDangerZone)
				require.WithinDuration(t, time.Now(), res.CheckedAt, 10*time.Second)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanupTestDB(t)

			if tt.setup != nil {
				tt.setup(t)
			}

			err := testRepoCoor.Check(ctx, tt.locCheck)

			if tt.wantErr {
				require.Error(t, err, "expected error but got nil")
			} else {
				require.NoError(t, err, "unexpected error: %v", err)
			}

			if tt.validate != nil {
				tt.validate(t, tt.locCheck)
			}
		})
	}
}

func TestCoordinatesRepository_GetStats(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping intergration test")
	}

	if testDB == nil {
		setupTestDB(t)
	}

	ctx := context.Background()

	tests := []struct {
		name              string
		timeWindowMinutes int
		setup             func(t *testing.T)
		wantErr           bool
		validate          func(t *testing.T, res []domain.ZoneStat)
	}{
		{
			name:              "success",
			timeWindowMinutes: 10,
			setup: func(t *testing.T) {
				incident := &domain.Incident{
					Title:       "Incident",
					Description: "Description",
					Lat:         50.0,
					Long:        50.0,
					Radius:      1000,
					Active:      true,
				}
				err := testRepo.Create(ctx, incident)
				require.NoError(t, err)

				for i := 1; i <= 6; i++ {
					locCheck := &domain.LocationCheck{
						UserID: fmt.Sprintf("colorvax-%d", i%3),
						Lat:    50,
						Long:   50.01,
					}
					err = testRepoCoor.Check(ctx, locCheck)
					require.NoError(t, err)
				}
			},
			validate: func(t *testing.T, res []domain.ZoneStat) {
				require.Len(t, res, 1)
				require.Equal(t, res[0].UserCount, 3)
				require.Equal(t, res[0].ZoneID, 1)
			},
		},
		{
			name:              "success - one user, several attempts",
			timeWindowMinutes: 10,
			setup: func(t *testing.T) {
				incident := &domain.Incident{
					Title:       "Incident",
					Description: "Description",
					Lat:         50.0,
					Long:        50.0,
					Radius:      1000,
					Active:      true,
				}
				err := testRepo.Create(ctx, incident)
				require.NoError(t, err)

				for i := 1; i <= 3; i++ {
					locCheck := &domain.LocationCheck{
						UserID: "colorvax",
						Lat:    50,
						Long:   50.01,
					}
					err = testRepoCoor.Check(ctx, locCheck)
					require.NoError(t, err)
				}
			},
			validate: func(t *testing.T, res []domain.ZoneStat) {
				require.Len(t, res, 1)
				require.Equal(t, res[0].UserCount, 1)
				require.Equal(t, res[0].ZoneID, 1)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cleanupTestDB(t)
		})

		if tt.setup != nil {
			tt.setup(t)
		}

		res, err := testRepoCoor.GetStats(ctx, tt.timeWindowMinutes)

		if tt.wantErr {
			require.Error(t, err, "expected error but got nil")
		} else {
			require.NoError(t, err, "unexpected error: %v", err)
		}

		if tt.validate != nil {
			tt.validate(t, res)
		}
	}
}
