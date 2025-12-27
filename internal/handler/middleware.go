package handler

import (
	"net/http"
	"red_collar/internal/service"
)

func apiKeyMiddleware(validAPIKey string, logger service.LoggerInterfaces) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			apiKey := r.Header.Get("X-API-Key")

			if apiKey != validAPIKey {
				logger.Error("invalid api key")
				writeAPIResponse(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid api key")
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
