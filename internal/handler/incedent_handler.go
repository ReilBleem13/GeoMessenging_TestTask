package handler

import (
	"encoding/json"
	"net/http"
	"red_collar/internal/domain"
	"red_collar/internal/service"

	"github.com/theartofdevel/logging"
)

// POST /api/v1/incidents
func (h *Handler) handleCreateIncedent(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var req newIncedentJSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", logging.ErrAttr(err))
		h.WriteError(w, domain.ErrInvalidRequest("invalid json payload"))
		return
	}

	in := service.CreateIncedentRequestInput{
		Title:       req.Title,
		Description: req.Description,
		Lat:         req.Lat,
		Long:        req.Long,
		Radius:      req.Radius,
		Active:      req.Active,
	}

	out, err := h.svc.CreateIncedent(r.Context(), &in)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	resp := incedentRequestResponse{
		Incendent: out,
	}
	writeJSON(w, http.StatusCreated, resp)
}

// GET /api/v1/incidents

// GET /api/v1/incidents{id}

// PUT /api/v1/incidents{id}

// DELETE /api/v1/incidents{id}
