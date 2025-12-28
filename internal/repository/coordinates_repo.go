package repository

import (
	"context"
	"database/sql"
	"red_collar/internal/domain"

	"github.com/jmoiron/sqlx"
)

type CoordinatesRepository struct {
	db *sqlx.DB
}

func NewCoordinatesRepository(db *sqlx.DB) *CoordinatesRepository {
	return &CoordinatesRepository{
		db: db,
	}
}

func (c *CoordinatesRepository) Check(ctx context.Context, locCheck *domain.LocationCheck) error {
	tx, err := c.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	checkQuery := `
		SELECT id 
		FROM incidents
		WHERE active = true 
			AND ST_DWithin(
					geom,
					ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography,
					radius_m
				)
		ORDER BY ST_Distance(
			geom,
			ST_SetSRID(ST_MakePoint($1, $2), 4326)::geography
		)
		LIMIT 1;
	`

	var incidentID int
	if err := tx.GetContext(ctx, &incidentID, checkQuery, locCheck.Long, locCheck.Lat); err != nil {
		if err != sql.ErrNoRows {
			return err
		}
	}

	if incidentID == 0 {
		locCheck.NearestID = nil
		locCheck.InDangerZone = false
	} else {
		locCheck.NearestID = &incidentID
		locCheck.InDangerZone = true
	}

	insertCheckQuery := `
		INSERT INTO location_checks (
			user_id, lat, long, in_danger_zone, nearest_id
		)
		VALUES($1, $2, $3, $4, $5)
		RETURNING id, checked_at
	`

	err = tx.QueryRowContext(ctx, insertCheckQuery,
		locCheck.UserID,
		locCheck.Lat,
		locCheck.Long,
		locCheck.InDangerZone,
		locCheck.NearestID,
	).Scan(&locCheck.ID, &locCheck.CheckedAt)
	if err != nil {
		return err
	}
	return tx.Commit()
}

func (c *CoordinatesRepository) GetStats(ctx context.Context, timeWindowMinutes int) ([]domain.ZoneStat, error) {
	getQuery := `
		SELECT 
			nearest_id as zone_id,
			COUNT(DISTINCT user_id) as user_count
		FROM location_checks
		WHERE in_danger_zone = true
			AND checked_at >= NOW() - INTERVAL '1 minute' * $1
		GROUP BY nearest_id
		ORDER BY zone_id
	`

	var stats []domain.ZoneStat
	err := c.db.SelectContext(ctx, &stats, getQuery, timeWindowMinutes)
	if err != nil {
		return nil, err
	}
	return stats, nil
}
