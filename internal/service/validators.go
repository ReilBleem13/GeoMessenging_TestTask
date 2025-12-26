package service

import (
	"red_collar/internal/domain"
	"strconv"
	"strings"
)

const (
	defaultLimit = 5
	maxLimit     = 50
)

func validateCreateIncidentInput(in *CreateIncidentRequestInput) error {
	if strings.TrimSpace(in.Title) == "" {
		return domain.ErrInvalidValidation("title is required")
	}

	if err := validateLatLong(in.Lat, in.Long); err != nil {
		return err
	}

	if in.Radius < 5 {
		return domain.ErrInvalidValidation("radius must be more than 5 meters")
	}
	return nil
}

func validateFullUpdateIncidentInput(in *FullUpdateIncidentRequestInput) (int, error) {
	id, err := validateID(in.ID)
	if err != nil {
		return 0, err
	}

	if strings.TrimSpace(in.Title) == "" {
		return 0, domain.ErrInvalidValidation("title is required")
	}

	if err := validateLatLong(in.Lat, in.Long); err != nil {
		return 0, err
	}

	if in.Radius < 5 {
		return 0, domain.ErrInvalidValidation("radius must be more than 5 meters")
	}
	return id, nil
}

func validateID(id string) (int, error) {
	if id == "" {
		return 0, domain.ErrInvalidValidation("id is required")
	}

	idInt, err := strconv.Atoi(id)
	if err != nil {
		return 0, domain.ErrInvalidValidation("invalid id format, must be integer")
	}
	return idInt, nil
}

func validateLatLong(lat, long float64) error {
	if lat < -90 || lat > 90 || long < -180 || long > 180 {
		return domain.ErrInvalidValidation("lat or long is invalid")
	}
	return nil
}

func validatePaginate(rawLimit, rawPage string) (int, int, int, error) {
	limit, err := strconv.Atoi(rawLimit)
	if err != nil {
		return 0, 0, 0, domain.ErrInvalidValidation("invalid limit format, must be integer")
	}

	if limit < 1 {
		limit = defaultLimit
	}

	if limit > maxLimit {
		limit = maxLimit
	}

	page, err := strconv.Atoi(rawPage)
	if err != nil {
		return 0, 0, 0, domain.ErrInvalidValidation("invalid page format, must be integer")
	}

	if page < 1 {
		page = 1
	}

	offset := (page - 1) * limit
	return offset, limit, page, nil
}
