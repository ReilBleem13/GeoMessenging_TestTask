package service

import (
	"context"
	"red_collar/internal/domain"

	"github.com/theartofdevel/logging"
)

func (s *Service) CreateIncident(ctx context.Context, in *CreateIncidentRequestInput) (*domain.Incident, error) {
	if err := validateCreateIncidentInput(in); err != nil {
		s.logger.Error("create incident request validation failed",
			logging.StringAttr("title", in.Title),
			logging.Float64Attr("lat", in.Lat),
			logging.Float64Attr("long", in.Long),
			logging.IntAttr("radius", in.Radius),
			logging.ErrAttr(err),
		)
		return nil, err
	}

	s.logger.Info("attempt to create incident",
		logging.StringAttr("title", in.Title),
	)

	incident := mapCreateIncidentInputToDomain(in)

	err := s.incidents.Create(ctx, incident)
	if err != nil {
		s.logger.Error("create incident request repository error",
			logging.StringAttr("title", in.Title),
			logging.ErrAttr(err),
		)
		return nil, err
	}

	s.logger.Info("incident was successfully created",
		logging.StringAttr("title", in.Title),
	)
	return incident, nil
}

func (s *Service) GetIncidentByID(ctx context.Context, rawID string) (*domain.Incident, error) {
	id, err := validateID(rawID)
	if err != nil {
		s.logger.Error("get incident by id validation failed",
			logging.StringAttr("id", rawID),
			logging.ErrAttr(err),
		)
		return nil, err
	}

	s.logger.Info("attempt to get incident by id",
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

	s.logger.Info("incident was successfully got",
		logging.IntAttr("id", id),
	)
	return incident, nil
}

func (s *Service) PaginateIncident(ctx context.Context, rawLimit, rawPage string) (*PaginateIncidentsOutput, error) {
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

	s.logger.Info("incidents was successfully paginated")
	return out, nil
}

func (s *Service) DeleteIncident(ctx context.Context, rawID string) error {
	id, err := validateID(rawID)
	if err != nil {
		s.logger.Error("delete incident validation failed",
			logging.StringAttr("id", rawID),
			logging.ErrAttr(err),
		)
		return err
	}

	s.logger.Info("attempt to delete",
		logging.IntAttr("id", id),
	)

	if err := s.incidents.Delete(ctx, id); err != nil {
		s.logger.Error("delete repository error",
			logging.IntAttr("id", id),
			logging.ErrAttr(err),
		)
		return err
	}

	s.logger.Info("incedent was successfully deleted",
		logging.IntAttr("id", id),
	)
	return nil
}

func (s *Service) FullUpdateIncident(ctx context.Context, in *FullUpdateIncidentRequestInput) (*domain.Incident, error) {
	id, err := validateFullUpdateIncidentInput(in)
	if err != nil {
		s.logger.Error("full update incident request validation failed",
			logging.StringAttr("id", in.ID),
			logging.StringAttr("title", in.Title),
			logging.Float64Attr("lat", in.Lat),
			logging.Float64Attr("long", in.Long),
			logging.IntAttr("radius", in.Radius),
			logging.ErrAttr(err),
		)
		return nil, err
	}

	s.logger.Info("attempt to full update incident",
		logging.StringAttr("id", in.ID),
	)

	incident := mapFullUpdateIncident(in, id)
	if err := s.incidents.FullUpdate(ctx, incident); err != nil {
		s.logger.Error("full update incident request repository error",
			logging.IntAttr("id", id),
			logging.ErrAttr(err),
		)
		return nil, err
	}

	s.logger.Info("incident was successfully full updated",
		logging.IntAttr("id", id),
	)
	return incident, nil
}
