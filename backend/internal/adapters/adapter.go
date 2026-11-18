// Package adapters implementa paneles especificos por tipo de contenedor (ej. Terraria:
// jugadores conectados). Ver docs/adapters.md para la guia de como agregar uno nuevo.
package adapters

import "time"

// ContainerInfo es lo minimo que un Adapter necesita para decidir si le aplica a un
// contenedor y para arrancar una instancia de seguimiento.
type ContainerInfo struct {
	ID     string
	Name   string
	Image  string
	Labels map[string]string
}

// Adapter se registra una vez por tipo de contenedor soportado (ver Register en
// registry.go, llamado desde el init() de cada paquete de adapter via blank import).
type Adapter interface {
	Slug() string
	Matches(c ContainerInfo) bool
	// Priority desempata cuando varios adapters matchean el mismo contenedor -- gana el
	// de Priority() mas alto. El adapter "generic" usa la prioridad minima posible.
	Priority() int
	NewInstance(c ContainerInfo) Instance
}

// Instance es el estado vivo de UN contenedor concreto mientras el Manager (manager.go)
// le va pasando lineas de log en tiempo real.
type Instance interface {
	// Snapshot devuelve el estado actual, serializable a JSON -- es lo que ve el frontend.
	Snapshot() any
	OnLogLine(line string, receivedAt time.Time)
	// Events emite eventos discretos (ej. "player_joined") para publicarlos por SSE
	// ademas del snapshot periodico. Puede devolver nil si el adapter no tiene eventos
	// discretos que emitir (el generic, por ejemplo).
	Events() <-chan Event
	Close()
}

// Event es un evento discreto de un adapter, publicado en el hub SSE como "adapter_event".
type Event struct {
	ContainerID string    `json:"container_id"`
	Adapter     string    `json:"adapter"`
	Type        string    `json:"type"`
	Data        any       `json:"data"`
	Time        time.Time `json:"time"`
}
