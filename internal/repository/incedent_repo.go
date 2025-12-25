package repository

import (
	"context"
	"database/sql"
	"red_collar/internal/domain"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type IncidentRepository struct {
	db *sqlx.DB
}

func NewIncidentRepository(db *sqlx.DB) *IncidentRepository {
	return &IncidentRepository{
		db: db,
	}
}

func (ip *IncidentRepository) Create(ctx context.Context, incedent *domain.Incident) error {
	createIncidentQuery := `
		INSERT INTO incidents (title, description, lat, long, radius_m, active)
		VALUES($1, $2, $3, $4, $5, $6)		
		RETURNING id, created_at, updated_at
	`

	err := ip.db.QueryRowContext(ctx, createIncidentQuery,
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

func (ip *IncidentRepository) GetByID(ctx context.Context, id int) (*domain.Incident, error) {
	getIncidentQuery := `
		SELECT
			id, title, description, lat, long, radius_m, active, created_at, updated_at  
		FROM incidents 
		WHERE id = $1
	`

	var incident domain.Incident

	if err := ip.db.GetContext(ctx, &incident, getIncidentQuery, id); err != nil {
		if err == sql.ErrNoRows {
			return nil, domain.ErrNotFound("incident is not exists")
		}
		return nil, err
	}
	return &incident, nil
}

func (ip *IncidentRepository) Paginate(ctx context.Context, limit, offset int) ([]domain.Incident, int, error) {
	var incidents []domain.Incident
	var total int

	totalQuery := `SELECT COUNT(*) FROM incidents`
	if err := ip.db.GetContext(ctx, &total, totalQuery); err != nil {
		return nil, 0, err
	}

	getQuery := `
		SELECT id, title, description, lat, long, radius_m, active, created_at, updated_at
		FROM incidents
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	err := ip.db.SelectContext(ctx, &incidents, getQuery, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return incidents, total, nil
}
