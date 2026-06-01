package handler

import (
	nethttp "net/http"

	"auction-live/backend/internal/monitoring"
)

type MetricsHandler struct {
	metrics *monitoring.Metrics
}

func NewMetricsHandler(metrics *monitoring.Metrics) *MetricsHandler {
	return &MetricsHandler{metrics: metrics}
}

func (h *MetricsHandler) GetMetrics(w nethttp.ResponseWriter, r *nethttp.Request) {
	if h.metrics == nil {
		w.WriteHeader(nethttp.StatusNoContent)
		return
	}

	w.Header().Set("Content-Type", "text/plain; version=0.0.4; charset=utf-8")
	w.WriteHeader(nethttp.StatusOK)
	_, _ = w.Write([]byte(h.metrics.RenderPrometheus()))
}
