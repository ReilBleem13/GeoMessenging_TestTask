package service

import "strings"

func validateCreateIncedentInput(in *CreateIncedentRequestInput) error {
	if strings.TrimSpace(in.Title) == "" {
		return ErrTitleRequired
	}

	if in.Lat < -90 || in.Lat > 90 || in.Long < -180 || in.Long > 180 {
		return ErrInvalidCoordinaates
	}

	if in.Radius < 5 {
		return ErrInvalidRadius
	}
	return nil
}
