// Package terraria detecta el servidor vanilla de Terraria (hexlo/terraria-server-docker,
// el que corre en este mismo homelab) y expone jugadores conectados + historial de
// entradas/salidas, parseando el log de la partida -- no hay TShock instalado, asi que
// no hay API REST que consultar, solo texto plano.
//
// Formato real observado en el log de este servidor (confirmado en vivo, sin comillas
// alrededor del nombre):
//
//	<nombre> has joined.
//	<nombre> has left.
package terraria

import (
	"regexp"
	"strings"
	"sync"
	"time"

	"homelab-dashboard/backend/internal/adapters"
)

func init() {
	adapters.Register(adapter{})
}

type adapter struct{}

func (adapter) Slug() string { return "terraria" }

func (adapter) Matches(c adapters.ContainerInfo) bool {
	return strings.Contains(c.Image, "hexlo/terraria-server-docker")
}

func (adapter) Priority() int { return 100 }

func (adapter) NewInstance(c adapters.ContainerInfo) adapters.Instance {
	return &instance{
		containerID: c.ID,
		online:      map[string]bool{},
		events:      make(chan adapters.Event, 16),
	}
}

var (
	joinRe  = regexp.MustCompile(`^(.+) has joined\.$`)
	leaveRe = regexp.MustCompile(`^(.+) has left\.$`)
)

const maxRecent = 30

type playerEvent struct {
	Name string    `json:"name"`
	Type string    `json:"type"` // "joined" | "left"
	Time time.Time `json:"time"`
}

type snapshot struct {
	Online []string      `json:"online"`
	Recent []playerEvent `json:"recent"`
}

type instance struct {
	containerID string

	mu     sync.Mutex
	online map[string]bool
	recent []playerEvent

	events chan adapters.Event
}

func (i *instance) OnLogLine(line string, receivedAt time.Time) {
	line = strings.TrimSpace(line)

	var name, typ string
	if m := joinRe.FindStringSubmatch(line); m != nil {
		name, typ = m[1], "joined"
	} else if m := leaveRe.FindStringSubmatch(line); m != nil {
		name, typ = m[1], "left"
	} else {
		return
	}

	ev := playerEvent{Name: name, Type: typ, Time: receivedAt}

	i.mu.Lock()
	if typ == "joined" {
		i.online[name] = true
	} else {
		delete(i.online, name)
	}
	i.recent = append(i.recent, ev)
	if len(i.recent) > maxRecent {
		i.recent = i.recent[len(i.recent)-maxRecent:]
	}
	i.mu.Unlock()

	select {
	case i.events <- adapters.Event{
		ContainerID: i.containerID,
		Adapter:     "terraria",
		Type:        "player_" + typ,
		Data:        ev,
		Time:        receivedAt,
	}:
	default:
		// nadie esta escuchando ahora mismo (nadie tiene el stream abierto) -- no bloquea
	}
}

func (i *instance) Snapshot() any {
	i.mu.Lock()
	defer i.mu.Unlock()

	online := make([]string, 0, len(i.online))
	for name := range i.online {
		online = append(online, name)
	}
	recent := make([]playerEvent, len(i.recent))
	copy(recent, i.recent)

	return snapshot{Online: online, Recent: recent}
}

func (i *instance) Events() <-chan adapters.Event { return i.events }

// Close es intencionalmente un no-op sobre el canal: cerrarlo aqui podria correr en
// paralelo con un ultimo OnLogLine todavia en vuelo (FollowLogs se cancela via context,
// no de forma sincronica) y esa combinacion es un panic clasico de Go ("send on closed
// channel"). El Manager deja de leer el canal via ctx.Done() en vez de depender de que
// se cierre -- el canal simplemente se vuelve inalcanzable y el GC se encarga.
func (i *instance) Close() {}
