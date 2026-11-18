// Package generic es el adapter de respaldo: matchea cualquier contenedor, no muestra
// nada especifico. Se registra con la prioridad minima posible para que cualquier
// adapter mas especifico (terraria, minecraft, ...) gane cuando tambien matchee.
package generic

import (
	"math"
	"time"

	"homelab-dashboard/backend/internal/adapters"
)

func init() {
	adapters.Register(adapter{})
}

type adapter struct{}

func (adapter) Slug() string                        { return "generic" }
func (adapter) Matches(adapters.ContainerInfo) bool { return true }
func (adapter) Priority() int                       { return math.MinInt32 }
func (adapter) NewInstance(adapters.ContainerInfo) adapters.Instance {
	return instance{}
}

// instance no guarda nada -- el contenedor ya se ve bien con la info generica que
// dockerapi.ContainerSummary expone (nombre, imagen, estado, stats, puertos).
type instance struct{}

func (instance) Snapshot() any                               { return nil }
func (instance) OnLogLine(line string, receivedAt time.Time) {}
func (instance) Events() <-chan adapters.Event               { return nil }
func (instance) Close()                                      {}
