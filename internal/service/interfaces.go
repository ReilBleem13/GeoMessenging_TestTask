package service

import (
	"context"
	"red_collar/internal/domain"
)

type IncedentRepositoryInterface interface {
	Create(ctx context.Context, incedent *domain.Incedent) error
}

type LoggerInterfaces interface {
	Debug(msg string, params ...any)
	Info(msg string, params ...any)
	Warn(msg string, params ...any)
	Error(msg string, params ...any)
}
