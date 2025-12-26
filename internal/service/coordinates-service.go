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
	return check, nil
}
