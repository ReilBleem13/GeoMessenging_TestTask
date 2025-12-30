package handler

import "red_collar/internal/domain"

// IncidentJSON представляет данные для создания/обновления инцидента
// @Description Данные инцидента (опасной зоны)
type IncidentJSON struct {
	Title       string  `json:"title"`
	Description *string `json:"description,omitempty"`
	Lat         float64 `json:"lat"`
	Long        float64 `json:"long"`
	Radius      int     `json:"radius_m"`
	Active      *bool   `json:"active,omitempty"`
}

// CheckJSON представляет данные для проверки координат
// @Description Координаты пользователя для проверки
type CheckJSON struct {
	UserID string  `json:"used_id"`
	Lat    float64 `json:"lat"`
	Long   float64 `json:"long"`
}

// Responses
type incedentRequestResponse struct {
	Incendent *domain.Incident `json:"Incedent"`
}

type statsRequestResponse struct {
	Stats []domain.ZoneStat `json:"Stats"`
}
