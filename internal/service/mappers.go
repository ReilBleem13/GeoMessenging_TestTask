package service

import "red_collar/internal/domain"

func mapCreateIncedentInputToDomain(in *CreateIncedentRequestInput) *domain.Incedent {
	desc := ""
	if in.Description != nil {
		desc = *in.Description
	}

	active := true
	if in.Active != nil {
		active = *in.Active
	}

	return &domain.Incedent{
		Title:       in.Title,
		Description: desc,
		Lat:         in.Lat,
		Long:        in.Long,
		Radius:      in.Radius,
		Active:      active,
	}
}

func mapDomainIncedentToDTO(in *domain.Incedent) IncedentRequestDTO {
	return IncedentRequestDTO{
		ID:     in.ID,
		Title:  in.Title,
		Lat:    in.Lat,
		Long:   in.Long,
		Radius: in.Radius,
		Active: in.Active,
	}
}
