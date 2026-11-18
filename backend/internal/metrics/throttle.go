package metrics

import "homelab-dashboard/backend/internal/hostctl"

type ThrottleStatus struct {
	Available bool   `json:"available"`
	Raw       string `json:"raw,omitempty"`
}

// ReadThrottleStatus pregunta a hostctl por vcgencmd get_throttled. Hasta que el
// daemon de hostctl se instale en el host (Fase 6), esto simplemente no esta
// disponible -- se degrada con gracia en vez de romper el resto de las metricas.
func ReadThrottleStatus(hc *hostctl.Client) ThrottleStatus {
	status, err := hc.GetThrottled()
	if err != nil {
		return ThrottleStatus{Available: false}
	}
	return ThrottleStatus{Available: true, Raw: status.Raw}
}
