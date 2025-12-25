package repository

import (
	"context"
	"red_collar/internal/domain"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type IncedentRepository struct {
	db *sqlx.DB
}

func NewIncedentRepository(db *sqlx.DB) *IncedentRepository {
	return &IncedentRepository{
		db: db,
	}
}

func (ip *IncedentRepository) Create(ctx context.Context, incedent *domain.Incedent) error {
	createIncedentQuery := `
		INSERT INTO incidents (title, description, lat, long, radius_m, active)
		VALUES($1, $2, $3, $4, $5, $6)		
		RETURNING id, created_at, updated_at
	`

	err := ip.db.QueryRowContext(ctx, createIncedentQuery,
		incedent.Title,
		incedent.Description,
		incedent.Lat,
		incedent.Long,
		incedent.Radius,
		incedent.Active,
	).Scan(&incedent.ID, &incedent.CreatedAt, &incedent.UpdatedAt)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return domain.ErrIncedentExists()
		}
		return err
	}
	return nil
}
