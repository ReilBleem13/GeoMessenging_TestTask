package service

import (
	"context"
	"encoding/json"
	"red_collar/internal/domain"

	"github.com/theartofdevel/logging"
)

const (
	cacheKeyIncidentID = "incidentID:"
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
		logging.IntAttr("ID", id),
	)

	key := cacheKeyIncidentID + rawID
	incident, err := s.getIncidentFromCache(ctx, key)
	if incident != nil {
		return incident, nil
	}

	incident, err = s.incidents.GetByID(ctx, id)
	if err != nil {
		s.logger.Error("get by id incident repository error",
			logging.IntAttr("id", id),
			logging.ErrAttr(err),
		)
		return nil, err
	}

	s.saveIncidentToCache(ctx, incident, key)

	s.logger.Info("incident was successfully got",
		logging.IntAttr("id", id),
	)
	return incident, nil
}

func (s *Service) PaginateIncident(ctx context.Context, rawLimit, rawPage string) (*PaginateIncidentsOutput, error) {
	offset, limit, page, err := validatePaginate(rawLimit, rawPage)
	if err != nil {
		s.logger.Error("paginate incidents validation failed",
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

	key := cacheKeyIncidentID + rawID
	s.deleteIncidenFromCache(ctx, key)

	if err := s.incidents.Delete(ctx, id); err != nil {
		s.logger.Error("delete repository error",
			logging.IntAttr("id", id),
			logging.ErrAttr(err),
		)
		return err
	}

	s.logger.Info("incident was successfully deleted",
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

	key := cacheKeyIncidentID + in.ID
	s.deleteIncidenFromCache(ctx, key)

	s.logger.Info("incident was successfully full updated",
		logging.IntAttr("id", id),
	)
	return incident, nil
}

// Кеширование
func (s *Service) getIncidentFromCache(ctx context.Context, key string) (*domain.Incident, error) {
	data, err := s.cache.Get(ctx, key)
	if err != nil {
		s.logger.Warn("failed to get incident from cache",
			logging.StringAttr("key", key),
			logging.ErrAttr(err),
		)
		return nil, nil
	}

	if data == nil {
		return nil, nil
	}

	var incident domain.Incident
	if err := json.Unmarshal(data, &incident); err != nil {
		s.logger.Error("failed to unmarshal incident from cache",
			logging.StringAttr("key", key),
			logging.ErrAttr(err),
		)
		return nil, nil
	}

	s.logger.Info("successfully got from cache", logging.IntAttr("incidentID", incident.ID))
	return &incident, nil
}

func (s *Service) saveIncidentToCache(ctx context.Context, incident *domain.Incident, key string) {
	data, err := json.Marshal(incident)
	if err != nil {
		s.logger.Error("failed to marshal incident for cache",
			logging.IntAttr("incidentID", incident.ID),
			logging.ErrAttr(err),
		)
		return
	}

	if err := s.cache.Save(ctx, data, key); err != nil {
		s.logger.Error("failed to save incident to cache",
			logging.IntAttr("incidentID", incident.ID),
			logging.ErrAttr(err),
		)
		return
	}
	s.logger.Info("successfully saved to cache", logging.IntAttr("incidentID", incident.ID))
}

func (s *Service) deleteIncidenFromCache(ctx context.Context, key string) {
	deleted, err := s.cache.Delete(ctx, key)
	if err != nil {
		s.logger.Error("failed to delete incident from cache",
			logging.StringAttr("key", key),
			logging.ErrAttr(err),
		)
	}

	if deleted {
		s.logger.Info("successfully deleted incident from cache",
			logging.StringAttr("key", key),
		)
	}
}
