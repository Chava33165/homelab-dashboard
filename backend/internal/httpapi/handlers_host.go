package httpapi

import (
	"log"
	"net/http"

	"homelab-dashboard/backend/internal/auth"
	"homelab-dashboard/backend/internal/hostctl"
)

type hostHandlers struct {
	hostctl *hostctl.Client
}

func (h *hostHandlers) reboot(w http.ResponseWriter, r *http.Request) {
	log.Printf("audit: REBOOT del host pedido desde %s", auth.ClientIP(r))
	if err := h.hostctl.Reboot(); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *hostHandlers) poweroff(w http.ResponseWriter, r *http.Request) {
	log.Printf("audit: POWEROFF del host pedido desde %s", auth.ClientIP(r))
	if err := h.hostctl.Poweroff(); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}
