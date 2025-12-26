package handler

import (
	"net/http"
	"red_collar/internal/service"
)

type Handler struct {
	svc    *service.Service
	logger service.LoggerInterfaces
}

func NewHandler(svc *service.Service, logger service.LoggerInterfaces) *Handler {
	return &Handler{
		svc:    svc,
		logger: logger,
	}
}

func NewRouter(svc *service.Service, logger service.LoggerInterfaces) *http.ServeMux {
	h := NewHandler(svc, logger)
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/incidents", h.handleCreateIncident)
	mux.HandleFunc("GET /api/v1/incidents/{id}", h.handleGetIncidentByID)
	mux.HandleFunc("GET /api/v1/incidents", h.handlePaginate)
	mux.HandleFunc("DELETE /api/v1/incidents/{id}", h.handleDeleteIncident)
	mux.HandleFunc("PUT /api/v1/incidents/{id}", h.handlePutIncident)

	mux.HandleFunc("POST /api/v1/location/check", h.handleCheckCoordinates)

	mux.HandleFunc("GET /api/v1/system/health", h.handleHealth)
	return mux
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func NewServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:    addr,
		Handler: handler,
	}
}
