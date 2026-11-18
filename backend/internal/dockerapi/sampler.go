package dockerapi

import (
	"context"
	"log"
	"time"

	"github.com/moby/moby/client"
	"homelab-dashboard/backend/internal/sse"
)

// tickInterval es mas largo que el de metrics.Sampler (2s) porque cada stats() implica
// ~1s de espera interna por contenedor (mitigado pidiendolas en paralelo en
// ListWithStats, pero de todos modos es mas trabajo que leer /proc).
const tickInterval = 5 * time.Second

// Sampler publica la lista de contenedores + sus stats en el hub compartido cada
// tickInterval, como evento "containers". Complementa a WatchEvents (que informa
// altas/bajas de inmediato); este sampler es el que trae CPU/RAM actualizados.
//
// annotate le agrega a la lista lo que no es de Docker (el adapter de cada contenedor)
// para que el snapshot tenga la misma forma que GET /api/containers.
type Sampler struct {
	cli      *client.Client
	hub      *sse.Hub
	annotate func([]ContainerWithStats) any
}

func NewSampler(cli *client.Client, hub *sse.Hub, annotate func([]ContainerWithStats) any) *Sampler {
	return &Sampler{cli: cli, hub: hub, annotate: annotate}
}

func (s *Sampler) Run(ctx context.Context) {
	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			list, err := ListWithStats(ctx, s.cli)
			if err != nil {
				log.Printf("dockerapi sampler: %v", err)
				continue
			}
			s.hub.Publish("containers", s.annotate(list))
		}
	}
}
