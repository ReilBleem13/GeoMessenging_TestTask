package handler

import "red_collar/internal/domain"

type IncidentJSON struct {
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	Lat         float64 `json:"lat"`
	Long        float64 `json:"long"`
	Radius      int     `json:"radius_m"`
	Active      *bool   `json:"active,omitempty"`
}

type CheckJSON struct {
	UserID string  `json:"used_id"`
	Lat    float64 `json:"lat"`
	Long   float64 `json:"long"`
}

// Responses
type incedentRequestResponse struct {
	Incendent *domain.Incident `json:"Incedent"`
}
