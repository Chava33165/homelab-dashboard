package adapters

import (
	"context"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/moby/moby/client"
	"homelab-dashboard/backend/internal/dockerapi"
	"homelab-dashboard/backend/internal/sse"
)

const reconcileInterval = 5 * time.Second

type tracked struct {
	instance Instance
	slug     string
	cancel   context.CancelFunc
}

// Manager mantiene una Instance viva por cada contenedor corriendo cuyo adapter
// resuelto no sea "generic" -- le sigue los logs en tiempo real (con backfill de las
// ultimas 500 lineas al arrancar) y publica sus eventos discretos en el hub SSE.
type Manager struct {
	cli *client.Client
	hub *sse.Hub

	mu      sync.Mutex
	tracked map[string]*tracked
}

func NewManager(cli *client.Client, hub *sse.Hub) *Manager {
	return &Manager{cli: cli, hub: hub, tracked: map[string]*tracked{}}
}

func (m *Manager) Run(ctx context.Context) {
	m.reconcile(ctx)

	ticker := time.NewTicker(reconcileInterval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			m.closeAll()
			return
		case <-ticker.C:
			m.reconcile(ctx)
		}
	}
}

func (m *Manager) reconcile(parent context.Context) {
	list, err := dockerapi.ListContainers(parent, m.cli)
	if err != nil {
		return
	}

	running := map[string]bool{}
	for _, c := range list {
		if c.State != "running" {
			continue
		}
		running[c.ID] = true

		info := ContainerInfo{ID: c.ID, Name: c.Name, Image: c.Image, Labels: c.Labels}
		a := Resolve(info)
		if a == nil || a.Slug() == "generic" {
			continue
		}

		m.mu.Lock()
		_, already := m.tracked[c.ID]
		m.mu.Unlock()
		if !already {
			m.startTracking(parent, info, a)
		}
	}

	m.mu.Lock()
	var toStop []string
	for id := range m.tracked {
		if !running[id] {
			toStop = append(toStop, id)
		}
	}
	m.mu.Unlock()
	for _, id := range toStop {
		m.stopTracking(id)
	}
}

func (m *Manager) startTracking(parent context.Context, info ContainerInfo, a Adapter) {
	ctx, cancel := context.WithCancel(parent)
	inst := a.NewInstance(info)

	m.mu.Lock()
	m.tracked[info.ID] = &tracked{instance: inst, slug: a.Slug(), cancel: cancel}
	m.mu.Unlock()

	tty, err := dockerapi.IsTTY(ctx, m.cli, info.ID)
	if err != nil {
		log.Printf("adapters: IsTTY %s: %v", info.Name, err)
	}

	// El docker-socket-proxy (haproxy) corta las conexiones largas cada ~10 minutos.
	// Sin reconectar, la instancia se quedaba congelada con el ultimo estado (ej.
	// "0 jugadores" aunque alguien entrara despues). Se reconecta sobre la MISMA
	// instancia: el backfill de 500 lineas va solo la primera vez, y despues se pide
	// desde el momento del corte para no perder ni duplicar lineas.
	go func() {
		tail, since := "500", ""
		for ctx.Err() == nil {
			err := dockerapi.FollowLogs(ctx, m.cli, info.ID, tty, tail, since, func(line string) {
				inst.OnLogLine(line, time.Now())
			})
			if ctx.Err() != nil {
				return
			}
			if err != nil {
				log.Printf("adapters: logs de %s: %v (reconectando)", info.Name, err)
			}
			tail, since = "all", strconv.FormatInt(time.Now().Unix(), 10)
			select {
			case <-ctx.Done():
				return
			case <-time.After(2 * time.Second):
			}
		}
	}()

	go func() {
		events := inst.Events()
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-events:
				if !ok {
					return
				}
				m.hub.Publish("adapter_event", ev)
			}
		}
	}()
}

func (m *Manager) stopTracking(id string) {
	m.mu.Lock()
	t, ok := m.tracked[id]
	if ok {
		delete(m.tracked, id)
	}
	m.mu.Unlock()
	if ok {
		t.cancel()
		t.instance.Close()
	}
}

func (m *Manager) closeAll() {
	m.mu.Lock()
	ids := make([]string, 0, len(m.tracked))
	for id := range m.tracked {
		ids = append(ids, id)
	}
	m.mu.Unlock()
	for _, id := range ids {
		m.stopTracking(id)
	}
}

// AdapterInfo devuelve el slug del adapter resuelto y su snapshot actual para un
// contenedor. Si no esta siendo trackeado (adapter generic, o el reconcile de 5s
// todavia no lo alcanza), devuelve slug="generic" y data=nil -- el frontend cae al
// panel generico en ese caso.
func (m *Manager) AdapterInfo(containerID string) (slug string, data any) {
	m.mu.Lock()
	t, ok := m.tracked[containerID]
	m.mu.Unlock()
	if !ok {
		return "generic", nil
	}
	return t.slug, t.instance.Snapshot()
}

// Container es lo que la API expone por contenedor: la info generica de Docker mas lo
// que haya resuelto el adapter. Vive aqui (no en dockerapi, que no puede depender de
// adapters) y la usan tanto GET /api/containers como el snapshot SSE "containers" --
// asi ambos mandan exactamente la misma forma y el frontend nunca recibe una lista
// sin "adapter".
type Container struct {
	dockerapi.ContainerWithStats
	Adapter     string `json:"adapter"`
	AdapterData any    `json:"adapter_data,omitempty"`
}

func (m *Manager) Annotate(list []dockerapi.ContainerWithStats) []Container {
	out := make([]Container, len(list))
	for i, c := range list {
		slug, data := m.AdapterInfo(c.ID)
		out[i] = Container{ContainerWithStats: c, Adapter: slug, AdapterData: data}
	}
	return out
}
