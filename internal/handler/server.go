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

	mux.HandleFunc("/api/v1/incidents", h.handleCreateIncedent)
	return mux
}

func NewServer(addr string, handler http.Handler) *http.Server {
	return &http.Server{
		Addr:    addr,
		Handler: handler,
	}
}
