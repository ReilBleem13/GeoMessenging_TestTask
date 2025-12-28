package handler

import (
	"encoding/json"
	"net/http"
	"red_collar/internal/domain"
	"red_collar/internal/service"

	"github.com/theartofdevel/logging"
)

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
