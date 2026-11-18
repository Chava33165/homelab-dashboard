package httpapi

import (
	"net/http"

	"homelab-dashboard/backend/internal/metrics"
	"homelab-dashboard/backend/internal/sse"
)

type metricsHandlers struct {
	hub *sse.Hub
}

func (h *metricsHandlers) stream(w http.ResponseWriter, r *http.Request) {
	h.hub.ServeHTTP(w, r)
}

// speedtest corre bajo demanda -- nunca se dispara automaticamente.
func (h *metricsHandlers) speedtest(w http.ResponseWriter, r *http.Request) {
	result, err := metrics.RunSpeedtest(r.Context())
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, result)
}

// hostProcesses solo lo pide el panel de RAM mientras esta abierto -- recorrer /proc
// es barato pero no tiene sentido hacerlo en cada tick del sampler.
func (h *metricsHandlers) hostProcesses(w http.ResponseWriter, r *http.Request) {
	procs, err := metrics.ReadHostProcesses(metrics.HostProcPath)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, procs)
}
