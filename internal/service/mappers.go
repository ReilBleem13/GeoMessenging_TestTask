package service

import "red_collar/internal/domain"

func mapCreateIncedentInputToDomain(in *CreateIncedentRequestInput) *domain.Incident {
	desc := ""
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

// func mapDomainIncedentToDTO(in *domain.Incedent) *CreateIncedentOutput {
// 	return &CreateIncedentOutput{
// 		ID:     in.ID,
// 		Title:  in.Title,
// 		Lat:    in.Lat,
// 		Long:   in.Long,
// 		Radius: in.Radius,
// 		Active: in.Active,
// 	}
// }
