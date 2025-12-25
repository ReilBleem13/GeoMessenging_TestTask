package service

import "red_collar/internal/domain"

// Input
type CreateIncedentRequestInput struct {
	Title       string
	Description *string
	Lat         float64
	Long        float64
	Radius      int
	Active      *bool
}

// OutPut
type Pagination struct {
	Total int `json:"total"`
	Page  int `json:"page"`
	Limit int `json:"limit"`
	Pages int `json:"pages"`
}

type PaginateIncidentsOutput struct {
	Incidents  []domain.Incident `json:"data"`
	Pagination *Pagination       `json:"pagination"`
}
