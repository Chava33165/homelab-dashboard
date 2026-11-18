package adapters

import "sort"

// LabelOverride: un contenedor con esta label fuerza que adapter usar, sin importar la
// imagen -- escape hatch para imagenes con nombres no matcheables por Matches().
const LabelOverride = "homelab.adapter"

var registered []Adapter

// Register lo llama el init() de cada paquete de adapter (generic, terraria, ...) via
// blank import -- ver cmd/dashboard/main.go. El paquete adapters en si NUNCA importa a
// sus subpaquetes (generic/terraria importan a adapters, no al reves), asi que quien
// hace los blank imports tiene que ser alguien de mas arriba en el arbol (main.go), o
// se forma un import cycle.
func Register(a Adapter) {
	registered = append(registered, a)
}

// Resolve determina que adapter usar para un contenedor:
//  1. Si el contenedor tiene la label LabelOverride, se usa ese adapter (si existe).
//  2. Si no, el adapter registrado con mayor Priority() cuyo Matches() de true.
//  3. Si ninguno matchea, cae al adapter "generic" (Priority minima, Matches siempre true).
//
// Devuelve nil solo si ni siquiera "generic" esta registrado (bug de wiring, no deberia
// pasar en producción -- ver main.go, que blank-importa adapters/generic).
func Resolve(c ContainerInfo) Adapter {
	if slug, ok := c.Labels[LabelOverride]; ok {
		for _, a := range registered {
			if a.Slug() == slug {
				return a
			}
		}
	}

	sorted := make([]Adapter, len(registered))
	copy(sorted, registered)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Priority() > sorted[j].Priority() })

	for _, a := range sorted {
		if a.Matches(c) {
			return a
		}
	}
	return nil
}
