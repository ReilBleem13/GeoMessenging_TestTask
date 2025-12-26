package service

import "red_collar/internal/domain"

func mapCreateIncidentInputToDomain(in *CreateIncidentRequestInput) *domain.Incident {
	desc := "without description"
	if in.Description != nil {
		desc = *in.Description
	}

	active := true
	if in.Active != nil {
		active = *in.Active
	}

	return &domain.Incident{
		Title:       in.Title,
		Description: desc,
		Lat:         in.Lat,
		Long:        in.Long,
		Radius:      in.Radius,
		Active:      active,
	}
}

func mapFullUpdateIncident(in *FullUpdateIncidentRequestInput, id int) *domain.Incident {
	desc := "without description"
	if in.Description != nil {
		desc = *in.Description
	}

	active := true
	if in.Active != nil {
		active = *in.Active
	}

	return &domain.Incident{
		ID:          id,
		Title:       in.Title,
		Description: desc,
		Lat:         in.Lat,
		Long:        in.Long,
		Radius:      in.Radius,
		Active:      active,
	}
}
