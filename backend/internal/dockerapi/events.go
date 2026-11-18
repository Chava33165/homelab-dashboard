package dockerapi

import (
	"context"
	"log"

	"github.com/moby/moby/api/types/events"
	"github.com/moby/moby/client"
	"homelab-dashboard/backend/internal/sse"
)

type DockerEvent struct {
	Action string `json:"action"`
	ID     string `json:"id"`
	Name   string `json:"name,omitempty"`
	Time   int64  `json:"time"`
}

// WatchEvents escucha los eventos de contenedores (altas/bajas/start/stop/etc) del
// daemon de Docker y los publica en el hub compartido como evento "docker_event" --
// asi el frontend detecta contenedores creados/borrados sin tener que hacer polling.
// Se reconecta solo si el stream de eventos se corta.
func WatchEvents(ctx context.Context, cli *client.Client, hub *sse.Hub) {
	filters := client.Filters{}
	filters.Add("type", string(events.ContainerEventType))

	for {
		if ctx.Err() != nil {
			return
		}

		result := cli.Events(ctx, client.EventsListOptions{Filters: filters})
		drained := drainEvents(ctx, result, hub)
		if !drained {
			return
		}
	}
}

// drainEvents consume un stream de eventos hasta que se corte o el contexto se cancele.
// Devuelve false si hay que dejar de reintentar (contexto cancelado).
func drainEvents(ctx context.Context, result client.EventsResult, hub *sse.Hub) bool {
	for {
		select {
		case <-ctx.Done():
			return false
		case err, ok := <-result.Err:
			if !ok {
				return true
			}
			if err != nil {
				log.Printf("docker events: %v", err)
			}
			return true
		case msg, ok := <-result.Messages:
			if !ok {
				return true
			}
			hub.Publish("docker_event", DockerEvent{
				Action: string(msg.Action),
				ID:     msg.Actor.ID,
				Name:   msg.Actor.Attributes["name"],
				Time:   msg.Time,
			})
		}
	}
}
