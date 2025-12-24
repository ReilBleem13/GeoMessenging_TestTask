package service

import (
	"context"

	"github.com/theartofdevel/logging"
)

func (s *Service) CreateIncedent(ctx context.Context, in *CreateIncedentRequestInput) (*CreateIncedentOutput, error) {
	if err := validateCreateIncedentInput(in); err != nil {
		s.logger.Error("create incedent request validation failed",
			logging.StringAttr("title", in.Title),
			logging.Float64Attr("lat", in.Lat),
			logging.Float64Attr("long", in.Long),
			logging.IntAttr("radius", in.Radius),
		)
		return nil, err
	}

	s.logger.Info("attempt to create incedent",
		logging.StringAttr("prID", in.Title),
	)

	incedent := mapCreateIncedentInputToDomain(in)

	err := s.incedents.Create(ctx, incedent)
	if err != nil {
		s.logger.Error("create incedent request repository error",
			logging.StringAttr("title", in.Title),
		)
		return nil, err
	}

	out := &CreateIncedentOutput{
		Incedent: mapDomainIncedentToDTO(incedent),
	}

	s.logger.Info("incedent was successfully created",
		logging.StringAttr("title", in.Title),
	)
	return out, nil
}
