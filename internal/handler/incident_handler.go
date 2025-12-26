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
	var req IncidentJSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", logging.ErrAttr(err))
		h.WriteError(w, domain.ErrInvalidRequest("invalid json payload"))
		return
	}

	in := service.CreateIncidentRequestInput{
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

	out, err := h.svc.PaginateIncident(r.Context(), rawlimit, rawPage)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	writeJSON(w, http.StatusOK, out)
}

// GET /api/v1/incidents{id}
func (h *Handler) handleGetIncidentByID(w http.ResponseWriter, r *http.Request) {
	rawID := r.PathValue("id")

	out, err := h.svc.GetIncidentByID(r.Context(), rawID)
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
func (h *Handler) handlePutIncident(w http.ResponseWriter, r *http.Request) {
	rawID := r.PathValue("id")

	var req IncidentJSON
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("invalid request body", logging.ErrAttr(err))
		h.WriteError(w, domain.ErrInvalidRequest("invalid json payload"))
		return
	}

	in := &service.FullUpdateIncidentRequestInput{
		ID:          rawID,
		Title:       req.Title,
		Description: req.Description,
		Lat:         req.Lat,
		Long:        req.Long,
		Radius:      req.Radius,
		Active:      req.Active,
	}

	out, err := h.svc.FullUpdateIncident(r.Context(), in)
	if err != nil {
		h.WriteError(w, err)
		return
	}

	resp := incedentRequestResponse{
		Incendent: out,
	}
	writeJSON(w, http.StatusOK, resp)
}

// DELETE /api/v1/incidents{id}
func (h *Handler) handleDeleteIncident(w http.ResponseWriter, r *http.Request) {
	rawID := r.PathValue("id")

	if err := h.svc.DeleteIncident(r.Context(), rawID); err != nil {
		h.WriteError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, nil)
}
