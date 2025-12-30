package handler

import (
	"fmt"
	"net/http"
	"red_collar/internal/config"
	"red_collar/internal/service"

	httpSwagger "github.com/swaggo/http-swagger"
)

type Handler struct {
	svc                 *service.Service
	logger              service.LoggerInterfaces
	statsTimeWindowMins int
}

func NewHandler(svc *service.Service, logger service.LoggerInterfaces, statsTimeWindowsMins int) *Handler {
	return &Handler{
		svc:                 svc,
		logger:              logger,
		statsTimeWindowMins: statsTimeWindowsMins,
	}
}

func NewRouter(svc *service.Service, logger service.LoggerInterfaces, cfg *config.Config) *http.ServeMux {
	h := NewHandler(svc, logger, cfg.App.StatsTimeWindowMins)
	mux := http.NewServeMux()

	apiKeyAuth := apiKeyMiddleware(cfg.App.APIKey, logger)

	mux.Handle("POST /api/v1/incidents", apiKeyAuth(http.HandlerFunc(h.handleCreateIncident)))
	mux.Handle("GET /api/v1/incidents/{id}", apiKeyAuth(http.HandlerFunc(h.handleGetIncidentByID)))
	mux.Handle("GET /api/v1/incidents", apiKeyAuth(http.HandlerFunc(h.handlePaginate)))
	mux.Handle("DELETE /api/v1/incidents/{id}", apiKeyAuth(http.HandlerFunc(h.handleDeleteIncident)))
	mux.Handle("PUT /api/v1/incidents/{id}", apiKeyAuth(http.HandlerFunc(h.handlePutIncident)))

	mux.HandleFunc("POST /api/v1/location/check", h.handleCheckCoordinates)
	mux.HandleFunc("GET /api/v1/incidents/stats", h.handleStats)

	mux.HandleFunc("GET /api/v1/system/health", h.handleHealth)

	mux.HandleFunc("GET /swagger/", httpSwagger.Handler(
		httpSwagger.URL(fmt.Sprintf("http://localhost:%s/swagger/doc.json", cfg.App.Port)),
	))
	return mux
}

// @Summary      Health Check
// @Description  Проверка работоспособности сервиса
// @Tags         system
// @Accept       json
// @Produce      json
// @Success      200  {object}  map[string]string
// @Router       /system/health [get]
func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func NewServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:    addr,
		Handler: handler,
	}
}
