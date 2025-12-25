package handler

import (
	"encoding/json"
	"net/http"
	"red_collar/internal/domain"
	"red_collar/internal/service"

	"github.com/theartofdevel/logging"
)

// POST /api/v1/incidents
func (h *Handler) handleCreateIncident(w http.ResponseWriter, r *http.Request) {
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

	out, err := h.svc.CreateIncident(r.Context(), &in)
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
func (h *Handler) handlePaginate(w http.ResponseWriter, r *http.Request) {
	rawPage := r.URL.Query().Get("page")
	rawlimit := r.URL.Query().Get("limit")

	out, err := h.svc.Paginate(r.Context(), rawlimit, rawPage)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, out)
}

// GET /api/v1/incidents{id}
func (h *Handler) handleGetIncidentByID(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	out, err := h.svc.GetIncidentByID(r.Context(), id)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	resp := incedentRequestResponse{
		Incendent: out,
	}
	writeJSON(w, http.StatusOK, resp)
}

// PUT /api/v1/incidents{id}

// DELETE /api/v1/incidents{id}
