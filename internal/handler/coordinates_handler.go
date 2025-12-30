package handler

import (
	"encoding/json"
	"net/http"
	"red_collar/internal/domain"
	"red_collar/internal/service"

	"github.com/theartofdevel/logging"
)

// @Summary      Проверка координат
// @Description  Проверяет, находится ли пользователь в опасной зоне. Если да, отправляет webhook-уведомление.
// @Tags         location
// @Accept       json
// @Produce      json
// @Param        coordinates  body      CheckJSON  true  "Координаты пользователя"
// @Success      200          {object}  domain.LocationCheck
// @Failure      400          {object}  badRequestErrorResponse
// @Failure      500          {object}  internalServerErrorResponse
// @Router       /location/check [post]
func (h *Handler) handleCheckCoordinates(w http.ResponseWriter, r *http.Request) {
	var req CheckJSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", logging.ErrAttr(err))
		h.WriteError(w, domain.ErrInvalidRequest("invalid json payload"))
		return
	}

	in := &service.CheckCoordinatesRequestInput{
		UserID: req.UserID,
		Lat:    req.Lat,
		Long:   req.Long,
	}

	out, err := h.svc.CheckCoordinates(r.Context(), in)
	if err != nil {
		h.WriteError(w, err)
		return
	}
	writeJSON(w, 200, out)
}

// @Summary      Статистика по зонам
// @Description  Получает статистику по количеству пользователей в каждой зоне за указанный временной период
// @Tags         incidents
// @Accept       json
// @Produce      json
// @Success      200  {object}  statsRequestResponse
// @Failure      500  {object}  internalServerErrorResponse
// @Router       /incidents/stats [get]
func (h *Handler) handleStats(w http.ResponseWriter, r *http.Request) {
	out, err := h.svc.GetStats(r.Context(), h.statsTimeWindowMins)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	zones := statsRequestResponse{
		Stats: out,
	}
	writeJSON(w, 200, zones)
}
