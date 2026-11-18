# Cómo agregar un adapter nuevo

Un adapter le da a un tipo de contenedor (Terraria, Minecraft, lo que sea) un panel
propio en la vista de detalle, en vez del genérico. Se resuelve así (ver
`backend/internal/adapters/registry.go`):

1. Si el contenedor tiene la label Docker `homelab.adapter=<slug>`, se usa ese adapter
   sin importar la imagen.
2. Si no, se recorren los adapters registrados por `Priority()` descendente y gana el
   primero cuyo `Matches()` devuelva `true`.
3. Si ninguno matchea, cae al adapter `generic` (siempre matchea, prioridad mínima).

El ejemplo real ya funcionando es el de Terraria
(`backend/internal/adapters/terraria/terraria.go`) — detecta la imagen
`hexlo/terraria-server-docker` y parsea las líneas de log `<nombre> has joined.` /
`<nombre> has left.` del server vanilla (no hay TShock instalado, así que es parseo de
texto plano, no API REST).

## Backend: implementar la interfaz

Dos interfaces en `backend/internal/adapters/adapter.go`:

```go
type Adapter interface {
    Slug() string
    Matches(c ContainerInfo) bool
    Priority() int
    NewInstance(c ContainerInfo) Instance
}

type Instance interface {
    Snapshot() any                         // lo que ve el frontend, serializable a JSON
    OnLogLine(line string, receivedAt time.Time)
    Events() <-chan Event                  // eventos discretos en vivo (join/leave, etc) -- puede ser nil
    Close()
}
```

El `Manager` (`backend/internal/adapters/manager.go`) es quien maneja Docker: cada ~5s
lista los contenedores corriendo, resuelve el adapter de cada uno, y si no es
`generic` y todavía no lo está siguiendo, abre un stream de logs (`docker logs --tail
500 -f` vía `dockerapi.FollowLogs`, que ya limpia códigos ANSI) y le va mandando cada
línea a `Instance.OnLogLine`. El adapter **nunca** toca Docker directamente, solo
recibe texto ya limpio.

### Ejemplo: Minecraft (`itzg/minecraft-server`)

Ese log trae líneas como:

```
[12:34:56] [Server thread/INFO]: jugador joined the game
[12:34:58] [Server thread/INFO]: jugador left the game
```

Paso a paso:

1. Crear `backend/internal/adapters/minecraft/minecraft.go`:

```go
package minecraft

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

func (adapter) Slug() string { return "minecraft" }
func (adapter) Matches(c adapters.ContainerInfo) bool {
    return strings.Contains(c.Image, "itzg/minecraft-server")
}
func (adapter) Priority() int { return 100 }
func (adapter) NewInstance(c adapters.ContainerInfo) adapters.Instance {
    return &instance{containerID: c.ID, online: map[string]bool{}, events: make(chan adapters.Event, 16)}
}

var (
    joinRe  = regexp.MustCompile(`\]: (\S+) joined the game$`)
    leaveRe = regexp.MustCompile(`\]: (\S+) left the game$`)
)

// ... misma forma que terraria.go: OnLogLine matchea las regex, actualiza el mapa
// "online" + un slice "recent" acotado, Snapshot() devuelve {online, recent},
// Events() manda un adapters.Event por cada join/leave sin bloquear si nadie escucha.
```

   (copiar la estructura completa de `terraria/terraria.go` es lo más rápido — el
   `OnLogLine`/`Snapshot`/`Events`/`Close` son casi idénticos, solo cambian las regex).

2. Registrarlo con un blank import en `backend/cmd/dashboard/main.go`:

```go
_ "homelab-dashboard/backend/internal/adapters/minecraft"
```

3. El campo `"adapter"` y `"adapter_data"` ya aparecen solos en `/api/containers` para
   cualquier contenedor que matchee (ver `backend/internal/httpapi/handlers_docker.go`,
   `containerResponse`) -- no hace falta tocar la API a mano.

## Frontend: el panel

1. Crear `frontend/src/lib/adapters/MinecraftPanel.svelte`. Recibe dos props:
   `data` (lo que devolvió `Snapshot()`, tipado como `any` porque cada adapter define su
   propia forma) y `containerId` (para filtrar los eventos en vivo que le
   correspondan). Ver `TerrariaPanel.svelte` como plantilla — el patrón es: sembrar el
   estado local con `data`, y suscribirse a `mainStream.on("adapter_event", ...)`
   filtrando por `ev.container_id === containerId && ev.adapter === "minecraft"` para
   actualizar al instante entre los snapshots periódicos (~5s).

2. Agregarlo al mapa en `frontend/src/lib/adapters/adapterRegistry.ts`:

```ts
import MinecraftPanel from "./MinecraftPanel.svelte";

export const adapterRegistry: Record<string, Component<any>> = {
  terraria: TerrariaPanel,
  minecraft: MinecraftPanel,
};
```

`ContainerDetail.svelte` ya busca `adapterRegistry[container.adapter]` solo y lo
renderiza si existe -- no hace falta tocar esa vista.

## Notas

- Sin base de datos: el historial se reconstruye leyendo `docker logs --tail 500` al
  arrancar el dashboard (o cuando el Manager detecta el contenedor por primera vez) --
  sobrevive reinicios del dashboard, se pierde solo si el propio contenedor del
  adapter se borra y se recrea (evento raro y aceptable de perder).
- Si 500 líneas no alcanzan a llegar hasta el último evento real (mucho ruido de por
  medio, ej. generación de mundo), el historial simplemente empieza vacío -- no es un
  bug, es una limitación consciente de no guardar nada en disco.
