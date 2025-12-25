package service

import (
	"context"
	"red_collar/internal/domain"

	"github.com/theartofdevel/logging"
)

func (s *Service) CreateIncident(ctx context.Context, in *CreateIncedentRequestInput) (*domain.Incident, error) {
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
		logging.StringAttr("title", in.Title),
	)

	incedent := mapCreateIncedentInputToDomain(in)

	err := s.incidents.Create(ctx, incedent)
	if err != nil {
		s.logger.Error("create incident request repository error",
			logging.StringAttr("title", in.Title),
		)
		return nil, err
	}

	s.logger.Info("incedent was successfully created",
		logging.StringAttr("title", in.Title),
	)
	return incedent, nil
}

func (s *Service) GetIncidentByID(ctx context.Context, rawID string) (*domain.Incident, error) {
	id, err := validateGetIncedentByID(rawID)
	if err != nil {
		s.logger.Error("get incident by id validation failed",
			logging.StringAttr("id", rawID),
			logging.ErrAttr(err),
		)
		return nil, err
	}

	s.logger.Info("attempt to get incedent by id",
		logging.StringAttr("ID", rawID),
	)

	incident, err := s.incidents.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("get by id incident repository error",
			logging.IntAttr("id", id),
			logging.ErrAttr(err),
		)
		return nil, err
	}

	s.logger.Info("incedent was successfully got",
		logging.IntAttr("id", id),
	)
	return incident, nil
}

func (s *Service) Paginate(ctx context.Context, rawLimit, rawPage string) (*PaginateIncidentsOutput, error) {
	offset, limit, page, err := validatePaginate(rawLimit, rawPage)
	if err != nil {
		s.logger.Error("paginate incidetns validation failed",
			logging.StringAttr("rawLimit", rawLimit),
			logging.StringAttr("rawPage", rawPage),
			logging.ErrAttr(err),
		)
		return nil, err
	}

	s.logger.Info("attempt to paginate",
		logging.IntAttr("limit", limit),
		logging.IntAttr("offset", offset),
	)

	incidents, total, err := s.incidents.Paginate(ctx, limit, offset)
	if err != nil {
		s.logger.Error("paginate repository error",
			logging.StringAttr("rawLimit", rawLimit),
			logging.StringAttr("rawPage", rawPage),
			logging.ErrAttr(err),
		)
		return nil, err
	}

	out := &PaginateIncidentsOutput{
		Incidents: incidents,
		Pagination: &Pagination{
			Total: total,
			Page:  page,
			Limit: limit,
			Pages: (total + limit - 1) / limit,
		},
	}

	return out, nil
}
