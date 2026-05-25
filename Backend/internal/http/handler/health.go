package handler

import (
	nethttp "net/http"

	api "auction-live/backend/internal/api"
	"auction-live/backend/internal/config"
)

type HealthHandler struct {
	cfg config.Config
}

func NewHealthHandler(cfg config.Config) *HealthHandler {
	return &HealthHandler{cfg: cfg}
}

func (h *HealthHandler) GetHealth(w nethttp.ResponseWriter, r *nethttp.Request) {
	api.Success(w, nethttp.StatusOK, map[string]any{
		"ok":          true,
		"service":     "auction-live-backend",
		"environment": h.cfg.AppEnv,
		"port":        h.cfg.AppPort,
	})
}
