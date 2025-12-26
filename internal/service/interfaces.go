package service

import (
	"context"
	"red_collar/internal/domain"
)

type IncidentRepositoryInterface interface {
	Create(ctx context.Context, incedent *domain.Incident) error
	GetByID(ctx context.Context, id int) (*domain.Incident, error)
	Paginate(ctx context.Context, limit, offset int) ([]domain.Incident, int, error)
	Delete(ctx context.Context, id int) error
	FullUpdate(ctx context.Context, incident *domain.Incident) error
}

type CoordinatesRepositoryInterface interface {
	Check(ctx context.Context, locCheck *domain.LocationCheck) error
}

type LoggerInterfaces interface {
	Debug(msg string, params ...any)
	Info(msg string, params ...any)
	Warn(msg string, params ...any)
	Error(msg string, params ...any)
}
