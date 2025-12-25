package handler

import "red_collar/internal/domain"

type newIncedentJSON struct {
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	Lat         float64 `json:"lat"`
	Long        float64 `json:"long"`
	Radius      int     `json:"radius_m"`
	Active      *bool   `json:"active,omitempty"`
}

type getIncedentJSON struct {
	ID int `json:"id"`
}

type changeIncedentJSON struct {
	Lat    float64 `json:"lat"`
	Long   float64 `json:"long"`
	Radius *int    `json:"radius_m"`
}

// Responses

type incedentRequestResponse struct {
	Incendent *domain.Incident `json:"Incedent"`
}
