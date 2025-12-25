package service

import (
	"red_collar/internal/domain"
	"strings"
)

func validateCreateIncedentInput(in *CreateIncedentRequestInput) error {
	if strings.TrimSpace(in.Title) == "" {
		return domain.ErrInvalidValidation("title is required")
	}

	if in.Lat < -90 || in.Lat > 90 || in.Long < -180 || in.Long > 180 {
		return domain.ErrInvalidValidation("lat or long is invalid")
	}

	if in.Radius < 5 {
		return domain.ErrInvalidValidation("radius must be more than 5 meters")
	}
	return nil
}
