package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"red_collar/internal/config"
	"red_collar/internal/domain"

	"github.com/theartofdevel/logging"
)

func main() {
	ctx := context.Background()
	cfg := config.Get()

	level := "info"
	if cfg.App.Mode == "debug" {
		level = "debug"
	}

	logger := logging.NewLogger(
		logging.WithIsJSON(level != "debug"),
		logging.WithAddSource(level != "debug"),
		logging.WithLevel(level),
	)

	ctx = logging.ContextWithLogger(ctx, logger)

	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var check domain.LocationCheck
		if err := json.NewDecoder(r.Body).Decode(&check); err != nil {
			logging.L(ctx).Error("Error decoding request", logging.ErrAttr(err))
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprintf(w, "Error: %v", err)
			return
		}

		logging.L(ctx).Info("Received webhook",
			logging.IntAttr("CheckID", check.ID),
			logging.StringAttr("UserID", check.UserID),
			logging.BoolAttr("InDangerZone", check.InDangerZone),
			logging.IntAttr("NearestID", *check.NearestID),
			logging.Float64Attr("Lat", check.Lat),
			logging.Float64Attr("Long", check.Long),
		)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{
			"message": "webhook received",
		})
	})

	logging.L(ctx).Info("Webhook server starting", logging.StringAttr("addr", ":9090"))
	if err := http.ListenAndServe(":9090", nil); err != nil {
		logging.L(ctx).Error("Webhook server failed", logging.ErrAttr(err))
		log.Fatal(err)
	}
}
