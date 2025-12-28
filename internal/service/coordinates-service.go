package service

import (
	"context"
	"red_collar/internal/domain"

	"github.com/theartofdevel/logging"
)

func (s *Service) CheckCoordinates(ctx context.Context, in *CheckCoordinatesRequestInput) (*domain.LocationCheck, error) {
	if err := validateLatLong(in.Lat, in.Long); err != nil {
		s.logger.Error("check coordinates request validation failed",
			logging.Float64Attr("lat", in.Lat),
			logging.Float64Attr("long", in.Long),
			logging.StringAttr("usedID", in.UserID),
			logging.ErrAttr(err),
		)
		return nil, err
	}

	s.logger.Info("attempt to check coordinates", logging.StringAttr("usedID", in.UserID))

	check := mapCheckInputToDomain(in)

	err := s.coordinates.Check(ctx, check)
	if err != nil {
		s.logger.Error("check coordinates request repository error",
			logging.StringAttr("userID", in.UserID),
		)
		return nil, err
	}

	s.logger.Info("location was successfully checked",
		logging.StringAttr("user", in.UserID),
	)

	if check.InDangerZone {
		if err := s.queue.Enqueue(ctx, check); err != nil {
			s.logger.Error("failed to enqueue webhook task",
				logging.StringAttr("userID", in.UserID),
				logging.ErrAttr(err),
			)
		} else {
			s.logger.Info("webhook task enqueued",
				logging.StringAttr("userID", in.UserID),
			)
		}
	}
	return check, nil
}

func (s *Service) GetStats(ctx context.Context, timeWindowMinutes int) ([]domain.ZoneStat, error) {
	s.logger.Info("attempt to get stats")

	zones, err := s.coordinates.GetStats(ctx, timeWindowMinutes)
	if err != nil {
		s.logger.Error("failed to get stat repository error",
			logging.ErrAttr(err),
		)
		return nil, err
	}

	s.logger.Info("successfully got stats")
	return zones, nil
}
