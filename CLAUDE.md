# CLAUDE.md -- homelab-dashboard

## Qué es esto

Dashboard web para monitorear y controlar el homelab que corre en una Raspberry Pi 5
(4GB, ARM64, Debian trixie, Docker 29.8). Ya hay otros servicios corriendo en este Pi
(`terraria-server`, `task-hub`) que se vieron afectados en el pasado por saturación de
RAM -- la estabilidad de memoria es un requisito real de este proyecto, no un nice-to-have.

El dashboard puede: ver métricas del host en tiempo real, listar/controlar contenedores
Docker (incluyendo reiniciarlos), mostrar paneles específicos por tipo de contenedor
(Terraria primero), y **reiniciar/apagar la propia Pi**. Por tener estas dos últimas
capacidades, el modelo de seguridad se trata con el mismo cuidado que un panel de admin
real, no como un side project descartable.

## Stack y por qué

- **Backend**: Go + SDK oficial de Docker (`github.com/docker/docker/client`). Habla
  por TCP con `docker-socket-proxy` -- **nunca** monta `/var/run/docker.sock` directo
  en el contenedor del dashboard.
- **Frontend**: Svelte + Vite, compilado a estáticos y embebido en el binario Go vía
  `//go:embed`. Una sola imagen final, sin servidor de frontend separado.
- **Tiempo real**: SSE, no WebSockets. Todo el tráfico en vivo (métricas, eventos
  Docker, logs) es servidor→cliente; las acciones (start/stop/restart) van por REST
  normal. Un único endpoint `/api/stream` multiplexado por tipo de evento, más un
  segundo endpoint aparte solo para logs de un contenedor -- así nunca se pasa de 2
  conexiones SSE simultáneas (relevante por el límite de 6 conexiones HTTP/1.1
  concurrentes por origen de los navegadores, hasta que llegue HTTP/2 con Caddy).
- **Sin base de datos.** Todo el estado vive en memoria del proceso Go. El historial de
  Terraria se reconstruye leyendo `docker logs --tail 500` al arrancar -- sobrevive
  reinicios del dashboard sin escribir nada a disco (bueno también para la vida de la
  SD/eMMC).
- **Build**: no hay Go ni Node instalados en el host -- todo pasa por Docker multi-stage
  (`node:alpine` → `golang:alpine` → imagen final mínima `distroless/static`/`scratch`).
  Compilar Node+Go en la misma Pi (4GB, ~1.5GB libres normalmente) es el punto de mayor
  riesgo de RAM de todo el proyecto -- se mide explícitamente en la Fase 4
  (`free -m`/`docker stats` durante el build). Si resulta inestable, el plan B es
  compilar fuera (o GitHub Actions) y publicar en `ghcr.io`, igual que ya hace `task-hub`.

## Métricas del host

Se leen directo de `/proc`, `/sys` y disco montados **read-only** dentro del contenedor
del dashboard (`/host/proc`, `/host/sys`, `/host/root`) -- nunca a través del proxy de
Docker, sin privilegios especiales para esto.

**Gotcha real encontrado en Fase 2, importante para no repetirlo**: `/proc/mounts` y
`/proc/net/dev` son symlinks a `/proc/self/mounts` y `/proc/self/net/dev` -- "self"
siempre resuelve al namespace de quien está leyendo, así que aunque `/proc` esté
bind-monteado desde el host, leer `/host/proc/mounts` o `/host/proc/net/dev` devuelve
el namespace de montaje/red del **propio contenedor del dashboard** (sus binds de
`/etc/hostname` etc., sus 3 interfaces `eth0/eth1/eth2` de las 3 redes Docker), no el
del host. Se confirmó viendo literalmente esos binds del contenedor aparecer como
"particiones de disco" en el stream de métricas. La solución: el compose usa `pid:
host` (comparte el namespace de PIDs del host) y el código lee `/host/proc/1/mounts` /
`/host/proc/1/net/dev` -- el PID 1 del namespace compartido es el init real del host,
así que su namespace de montaje/red sí es el del host. Esto amplía lo que el
contenedor puede *ver* (lista de procesos del host), no lo que puede *tocar* -- sigue
sin capabilities, sin escritura, sin acceso a `docker.sock`.

- Temperatura: `/sys/class/thermal/thermal_zone0/temp` (sysfs plano, sin vcgencmd).
- RAM/swap: este Pi **ya tiene zram activo** (`/dev/zram0`, confirmado vía `/proc/swaps`
  en el host) -- no es algo pendiente de Fase 6, ya está funcionando. `docs/stability.md`
  documentará cómo verificarlo/ajustarlo, no cómo instalarlo desde cero.
- Latencia de red: **TCP connect cronometrado a 1.1.1.1:443**, no ICMP real. Decisión
  consciente: ICMP real pediría `cap_add: NET_RAW` (o tocar `net.ipv4.ping_group_range`
  del host para el modo sin privilegios) -- el connect TCP mide un proxy razonable de
  salud de red sin ampliar la superficie del contenedor. Documentado en `docs/security.md`.
- Speedtest on-demand (nunca automático): librería Go pura `github.com/showwin/speedtest-go`,
  sin shell-out a un binario externo -- mantiene la imagen final mínima.
- Throttling/voltaje (`vcgencmd get_throttled`): **no** se lee desde el contenedor del
  dashboard. Pasa por `hostctl` (ver abajo) porque `/dev/vcio` es `root:root` y no
  queremos correr el contenedor como root para esto.

## `hostctl`: el helper de host

Componente **separado**, fuera de Docker, corriendo directo en el host vía systemd
socket activation. Existe porque dos operaciones (`reboot`/`poweroff` y
`vcgencmd get_throttled`) requieren privilegios que el contenedor del dashboard
deliberadamente no tiene.

- Binario Go independiente (stdlib-only, su propio `go.mod` en `hostctl/`).
- `hostctl.socket`: crea `/run/homelab-hostctl.sock` con `SocketMode=0660`,
  `SocketGroup=homelab-hostctl` -- permisos correctos desde antes de que el proceso
  arranque.
- `hostctl.service`: corre como usuario dedicado `hostctl`, no root. Recibe el fd vía
  `LISTEN_FDS`. Endurecido con `NoNewPrivileges`, `ProtectSystem=strict`,
  `ProtectHome`, `PrivateTmp`, `RestrictSUIDSGID`, `MemoryDenyWriteExecute`.
- Protocolo: JSON-lines, **4 comandos fijos, sin argumentos libres**: `ping`,
  `get_throttled`, `reboot`, `poweroff`. Nunca hay `exec.Command` con input del cliente.
- `get_throttled`: regla `udev` pone `/dev/vcio` en un grupo `vcio`, y el usuario
  `hostctl` está en ese grupo -- sin volverlo root.
- `reboot`/`poweroff`: vía D-Bus (`org.freedesktop.login1.Manager`), autorizado por una
  regla **polkit** dedicada que solo le da esas dos acciones al usuario `hostctl`. Nada
  de `CAP_SYS_BOOT`. Requiere `policykit-1` instalado en el host (no siempre viene en
  instalaciones headless mínimas -- verificar en Fase 6).
- El contenedor `dashboard` llega al socket vía `group_add` (mismo gid
  `homelab-hostctl`) **más** un token compartido (`openssl rand -hex 32`) que debe
  mandar en cada request. Esta capa extra existe porque quien realmente "habla" con el
  socket es el proceso del contenedor, no el humano ya autenticado en la app web --
  y ambos pueden divergir si hay un bug o un RCE en el dashboard.

**Riesgo si el contenedor `dashboard` se compromete** (ver `docs/security.md` para el
detalle completo): contra Docker, solo puede hacer lo que el proxy permite (list/logs/
stats/start/stop/restart -- nada de `exec`, `build`, montar volúmenes nuevos, ni
escapar al host vía Docker). Contra `hostctl`, el peor caso con el token robado es
reboot/poweroff (DoS recuperable físicamente, sin pérdida de datos) y fuga de estado de
throttling (bajo impacto). La superficie de `hostctl` es intencionalmente mínima por
diseño. Todo comando aceptado se audita (`journalctl -u hostctl` + log del dashboard
con IP de origen).

## Docker: `docker-socket-proxy`

Nunca se monta `/var/run/docker.sock` en el contenedor del dashboard. El proxy
(`tecnativa/docker-socket-proxy`) corre en una red Docker `internal: true` (sin salida
a internet, inalcanzable desde fuera del propio dashboard) con permisos mínimos --
config exacta y razonada en `deploy/docker-socket-proxy.env`:
`CONTAINERS`, `EVENTS`, `PING`, `VERSION`, `POST` + `ALLOW_START`/`ALLOW_STOP`/
`ALLOW_RESTARTS`. Todo lo demás (`IMAGES`, `NETWORKS`, `VOLUMES`, `EXEC`, `BUILD`,
`COMMIT`, `SECRETS`, `SWARM`, etc.) queda en `0`.

## Aislamiento de red Docker

`cloudflared` necesita salida a internet para conectar al borde de Cloudflare, así que
no puede compartir la red `internal: true` del proxy de Docker. El contenedor
`dashboard` es multi-red:

- `proxy-net` (`internal: true`) -- con `docker-socket-proxy`. Sin salida a internet.
- `dashboard-net` (`internal: true`) -- con `caddy`, vía Tailscale. Sin salida a internet.
- `cloudflare-net` (red normal, con NAT) -- con `cloudflared`.

`docker-socket-proxy` **nunca** toca `cloudflare-net` -- queda inalcanzable incluso si
la vía pública (`cloudflared`) tuviera un problema.

## Acceso: dos vías

1. **Tailscale/LAN (privada)** -- igual que `terraria-server`, puertos publicados solo
   en la IP de Tailscale del Pi, nunca en `0.0.0.0`. HTTPS vía `tailscale cert`
   (certificado real de Let's Encrypt para el hostname `*.ts.net`, confiado
   automáticamente por Safari/iOS, sin instalar ningún perfil de CA) servido por Caddy.
   Se descartó una CA interna de Caddy por la fricción real de instalar/confiar un
   perfil en iOS. Renovación no automática -- timer propio en
   `deploy/systemd/tailscale-cert-renew.timer`.
2. **Cloudflare Tunnel + dominio propio (pública)** -- mismo patrón que ya usa
   `task-hub`. **Decisión explícita del usuario**: dado que este dashboard puede
   reiniciar/apagar la Pi y controlar todos los contenedores, lleva **Cloudflare Access
   obligatorio delante del túnel** (login de Cloudflare -- email OTP o Google/GitHub --
   antes de que el tráfico llegue siquiera al contenedor). Es una capa adicional, no un
   sustituto del login propio de la app. Cloudflare termina el TLS en su borde
   automáticamente, no pasa por Caddy.

## Login y sesión

Usuario único, hash bcrypt por env var (`ADMIN_PASSWORD_BCRYPT_HASH`), cookie de
sesión `HttpOnly` + `SameSite=Strict`, rate limiting en memoria por IP. El flag
`COOKIE_SECURE` es configurable por env var: `false` durante fases 2-5 (HTTP plano por
LAN/Tailscale, sin HTTPS todavía), pasa a `true` en Fase 6 cuando llega Caddy.

CSRF: `SameSite=Strict` + header custom obligatorio en mutaciones. Suficiente para este
caso -- un solo usuario, sin superficie multiusuario ni CORS habilitado hacia otros
orígenes.

**Gotcha real encontrado en Fase 2**: `ADMIN_PASSWORD_BCRYPT_HASH` y `SESSION_SECRET`
viven en `.env.secrets` (vía `env_file:`), no en `.env`. Docker Compose interpola `$`
incluso dentro de archivos usados con `env_file:` (no es exclusivo de `${...}`) -- un
hash bcrypt trae varios `$` (`$2a$14$...`) y sin escapar se corrompe en silencio
(confirmado empíricamente arrancando un contenedor real y volcando su entorno, no solo
con `docker compose config`). Usar siempre `scripts/hash-password.sh` (o
`make hash-password PASSWORD=...`), que genera el hash ya escapado (`$$` por cada `$`).

**Cambio de contraseña desde la propia UI** (post-Fase 6): `ADMIN_PASSWORD_BCRYPT_HASH`
ya solo es el hash de *bootstrap* -- la fuente de verdad real es
`internal/auth/password.go` (`PasswordStore`), que en el primer arranque copia ese
bootstrap a un archivo (`DataDir/admin_password_hash`, `DataDir` = `/data` por
defecto) y de ahí en adelante lee/escribe ese archivo. `POST /api/auth/change-password`
(autenticado, valida la contraseña actual, hashea la nueva con bcrypt ahí mismo) permite
cambiarla desde Ajustes sin recrear el contenedor ni volver a tocar `.env.secrets` --
justo lo que pidió el usuario ("que sola se hashee, sin mandarte la contraseña a ti").
Es la **única excepción** al principio de "sin base de datos" de este proyecto: a
diferencia del historial de Terraria (reconstruible desde los logs), una contraseña
*tiene* que persistir de verdad o el usuario pierde acceso. Por eso el compose monta un
volumen con nombre `dashboard-data:/data` (no bind mount) en vez de tmpfs. Como la
imagen final es `distroless` (sin shell, no hay `mkdir`/`chown` posibles ahí), el
directorio `/data` se crea con el dueño correcto (`10001:10001`, mismo UID/GID que
corre el proceso) en la etapa de build de Go (que sí tiene shell) y se copia con
`COPY --chown` a la etapa final -- Docker copia ese contenido+permisos dentro del
volumen nombrado la primera vez que está vacío, así el proceso sin privilegios puede
escribir ahí. Verificado en vivo: `ls -la` dentro del volumen muestra
`admin_password_hash` con dueño `10001:10001` y permisos `0600`; probado el flujo
completo (login → cambiar con password actual incorrecta → falla; cambiar con la
correcta → password vieja deja de servir, la nueva funciona de inmediato sin reiniciar
el contenedor).

## Sistema de adaptadores

```go
type Adapter interface {
    Slug() string
    Matches(c ContainerInfo) bool
    Priority() int
    NewInstance(c ContainerInfo) Instance
}

type Instance interface {
    Snapshot() any
    OnLogLine(line string, receivedAt time.Time)
    Events() <-chan Event
    Close()
}
```

**Nota**: `NewInstance` terminó sin `ctx`/`logs` (a diferencia de lo esbozado en Fase
1) -- quien abre el stream de logs de Docker y se lo va pasando línea por línea es el
`Manager` (`internal/adapters/manager.go`), no cada adapter. Así un adapter nunca
importa `dockerapi` ni toca la conexión al proxy directamente, solo recibe texto ya
limpio de ANSI vía `OnLogLine`. El `Manager` reconcilia cada ~5s: lista contenedores
corriendo, resuelve adapter, y si no es `generic` y no lo está siguiendo aún, abre
`dockerapi.FollowLogs(..., tail:"500", follow:true)` -- backfill + en vivo en una sola
llamada -- y publica `Instance.Events()` en el hub SSE como `adapter_event`.

Resolución: 1) label explícito `homelab.adapter=<slug>` en el contenedor si existe,
2) si no, el adapter registrado (vía blank import **desde `main.go`**, no desde
`registry.go` -- `generic` y `terraria` importan a `adapters`, así que si `adapters`
los importara a ellos se formaría un ciclo) con `Priority()` más alto cuyo `Matches()`
sea `true`, 3) si ninguno matchea, cae al adapter `generic` (`Priority()` mínimo,
`Matches()` siempre `true`).

El adapter de Terraria (`internal/adapters/terraria/`) detecta la imagen
`hexlo/terraria-server-docker` y parsea las líneas `<nombre> has joined.` /
`<nombre> has left.` del log vanilla -- **sin comillas alrededor del nombre**,
confirmado con logs reales de este mismo servidor (no hay TShock instalado, así que es
parseo de texto plano, no API REST). El log del vanilla server también mete códigos de
control ANSI a veces (`dockerapi.stripANSI`, agregado en Fase 4, ya los limpia antes de
que el adapter vea la línea).

`/api/containers` expone `adapter` (slug) y `adapter_data` (el `Snapshot()` del
Manager, o `null` si es `generic`) por contenedor -- la composición vive en
`internal/httpapi/handlers_docker.go` (`containerResponse`), no en `dockerapi`, por la
misma razón del ciclo de imports de arriba.

Cómo agregar un adapter nuevo: ver `docs/adapters.md` (ejemplo completo con Minecraft).

## Mem limits y estabilidad

Todos los servicios usan `mem_limit` (no `deploy.resources.limits` -- confirmado que
`docker compose up` fuera de Swarm ignora ese bloque). Valores actuales son
placeholders a medir con datos reales conforme avanzan las fases, mismo criterio que
se usó con `terraria-server` (medir con `docker stats`/`/proc` antes de fijar un
límite definitivo, no adivinar).

**Hallazgo crítico de Fase 2 (resuelto)**: por defecto, el firmware/bootloader del Pi 5
inyecta `cgroup_disable=memory` en la línea de comandos del kernel **antes** de
`cmdline.txt` -- no aparece en ningún archivo editable, así que no basta con buscarlo.
Docker avisaba `"Your kernel does not support memory limit capabilities"` y `mem_limit`
no protegía nada de verdad. Arreglo aplicado:
```
sudo sed -i 's/$/ cgroup_enable=memory cgroup_memory=1/' /boot/firmware/cmdline.txt
sudo reboot
```
**Ojo al verificar**: este sistema usa cgroup v2 (`docker info` → `Cgroup Version: 2`).
`/proc/cgroups` es la vista legacy de v1 y puede seguir mostrando el controlador
`memory` ausente **aunque el fix ya haya funcionado** -- es un falso negativo. El
archivo correcto para confirmar en v2 es:
```
cat /sys/fs/cgroup/cgroup.controllers   # debe incluir "memory"
```
Y para confirmar que un contenedor específico realmente tiene el límite aplicado (no
solo configurado):
```
docker inspect <contenedor> --format '{{.HostConfig.Memory}}'   # bytes configurados
cat /sys/fs/cgroup/system.slice/docker-<id-completo>.scope/memory.max   # limite real en el cgroup
```
Confirmado en este Pi: `dashboard` → 268435456 (256MB), `docker-socket-proxy` →
67108864 (64MB). Nota aparte: recrear (no solo reiniciar) los contenedores que ya
estaban corriendo antes del reboot es necesario para que tomen el límite -- Docker no
lo aplica retroactivamente a un contenedor que ya existía cuando el cgroup no estaba
disponible.

`docs/stability.md` (Fase 6) documenta zram, earlyoom, el watchdog de hardware del
Pi 5, y cómo diagnosticar un cuelgue por OOM después del hecho
(`journalctl -b -1`, revisar el OOM killer). `docs/retrofitting-mem-limits.md` (Fase 6)
da los `mem_limit` concretos recomendados para `task-hub` y `terraria-server`
(terraria midió ~1GB con 2 jugadores en este mismo Pi -- ver historial del proyecto).

## Estructura de carpetas

Ver el árbol completo en `README.md` / el plan de Fase 1. Resumen: `backend/` (Go,
`cmd/dashboard` + `internal/{config,auth,httpapi,sse,metrics,dockerapi,adapters,
hostctl,web}`), `frontend/` (Svelte+Vite), `hostctl/` (binario de host, systemd/polkit/
udev), `deploy/` (Caddyfile, env del proxy, timer de renovación de cert), `docs/`.

## Fases (checklist)

- [x] **Fase 1** -- Arquitectura y estructura de carpetas.
- [x] **Fase 2** -- Backend: métricas del host + login. Verificado en vivo en el Pi:
      login/logout/rate-limit/cookie/CSRF/401/stream SSE, todo correcto. Cgroup de
      memoria habilitado y `mem_limit` confirmado real a nivel cgroup (256MB dashboard,
      64MB proxy) -- ver "Estabilidad" más abajo.
- [x] **Fase 3** -- Integración Docker. Verificado en vivo en el Pi: `/api/containers`
      lista los 5 contenedores reales (dashboard, docker-socket-proxy, terraria-server,
      task-hub-backend, task-hub-cloudflared) con CPU/RAM reales; creación/borrado de un
      contenedor de prueba se reflejó al instante vía `docker_event` en el stream;
      logs en vivo, start/stop/restart probados contra un contenedor real (confirmado
      con `docker inspect`, no solo por el código HTTP); `/images/json` contra el proxy
      da 403 (IMAGES=0), `/containers/json` da 200 (CONTAINERS=1) -- el scoping de
      permisos sí restringe de verdad.
- [x] **Fase 4** -- Frontend completo (Svelte 5 + Vite + TS, diseño glass, dark/light,
      sparklines, View Transitions). **Rediseño posterior del usuario** ("Spider-Verse
      Gwen": paleta neón rosa/menta, temas claro/spider/oscuro via `LiquidSwitcher`,
      halftone comic, `chromatic-pop`) -- revisado a fondo: 0 errores de `svelte-check`,
      build limpio, toda la lógica (streams SSE, adaptadores, control de contenedores)
      intacta, verificado visualmente en las 4 vistas con datos reales. Tipografía
      final: **DM Sans + Outfit, autohospedadas** (`frontend/fonts/`) -- el rediseño
      las había puesto por CDN de Google Fonts, se corrigió para cumplir el requisito
      original de "sin CDN" (confirmado con Playwright: cero peticiones a
      fonts.googleapis.com/gstatic.com, las fuentes cargan desde `/assets/` propio).
      `LiquidSwitcher.svelte` combina `backdrop-filter` con un filtro SVG
      (`url(#switcher)`) con fallback `-webkit-backdrop-filter` sin el filtro para
      Safari -- el usuario confirmó en un iPhone real que se ve bien. Build
      multi-stage real verificado (`docker compose build`, node:22-alpine → golang:1.26
      → distroless). Verificado visualmente con Playwright en contenedor (sin Node/
      browser en el host): login, dashboard con métricas en vivo, detalle de contenedor
      con logs en vivo, tema oscuro/claro -- capturas revisadas, 0 errores de consola,
      sin overflow horizontal. Dos bugs reales encontrados y arreglados en el camino:
      (1) el frontend nunca llamaba a `mainStream.connect()`, así que el `EventSource`
      jamás se abría -- las métricas se quedaban en "Conectando..." para siempre aunque
      el backend funcionara perfecto (los contenedores sí cargaban porque usan un fetch
      REST aparte); (2) los logs de contenedores con TTY (como `terraria-server`) traían
      códigos de control ANSI del terminal sin limpiar -- se agregó `stripANSI()` en
      `internal/dockerapi/logs.go`.
- [x] **Fase 5** -- Adaptador de Terraria. Verificado en vivo: `/api/containers`
      resuelve `adapter: "terraria"` para `terraria-server` (imagen matchea), el panel
      del frontend renderiza "0 jugadores / Nadie conectado ahora." con datos reales
      del contenedor. El historial (`recent`) sale vacío ahora mismo -- confirmado que
      es correcto, no un bug: los últimos joins/leaves reales quedaron a 801 líneas del
      final del log (ruido de "Settling liquids" de las recargas de mundo empuja todo
      hacia atrás), fuera de la ventana de backfill de 500 líneas. Reinicio del
      dashboard confirmado sin errores, el Manager se re-engancha solo. **Pendiente de
      confirmar con un jugador real** (no se puede simular un handshake de Terraria
      real desde este Pi sin un cliente del juego): conectar/desconectar y ver
      join/leave aparecer en vivo en el panel sin refrescar.
- [~] **Fase 6** -- Todo el código/config listo y verificado en lo que no requiere
      `sudo` interactivo; dos pasos de instalación en el host quedaron pendientes de
      que el usuario los corra (no se pueden automatizar desde aquí):
      - **PWA**: manifest, íconos (generados con Playwright rasterizando un SVG, no
        había ImageMagick/rsvg-convert disponibles), service worker (JS plano en
        `public/`, no bundleado -- cache-first en `/assets/*` por el hash de Vite,
        nunca cachea `/api/*`). Verificado sirviendo con content-type correcto.
      - **HTTPS (Tailscale cert + Caddy)**: `Caddyfile`, compose de `caddy` ya activo
        (sin `profiles`), timer de renovación listo. **Pendiente**: `tailscale cert`
        pide `sudo` (o `sudo tailscale set --operator=$USER` una vez) -- no se pudo
        confirmar que tomara efecto en esta sesión. Ver `docs/access.md` para los
        comandos exactos.
      - **`hostctl`**: binario compilado y verificado (`make hostctl-build` funciona,
        arreglé un bug real del Dockerfile -- `docker create` fallaba con "no command
        specified" sin un `CMD` en el stage `scratch` final). Backend ya expone
        `/api/host/reboot` y `/api/host/poweroff`, probados en vivo: fallan con un
        error claro (`dial unix ... no such file or directory`) mientras `hostctl` no
        esté instalado -- confirmado visualmente que la UI (doble confirmación +
        mensaje de error) funciona perfecto en ese estado. **Pendiente**: instalación
        en el host (usuarios, grupos, systemd, polkit, udev -- todo requiere `sudo`,
        ver `hostctl/README.md` paso a paso).
      - **Cloudflare Tunnel + Access**: intencionalmente NO activado -- el usuario
        pidió quedarse por ahora solo en LAN + Tailscale. Queda detrás de
        `profiles: ["public"]`, documentado en `docs/access.md` para cuando se quiera.
      - **mem_limit retroactivo**: aplicado de verdad a `terraria-server` (2048m) y
        `task-hub` (256m backend, 128m cloudflared), verificado a nivel cgroup real.
      - **Guía de estabilidad**: `docs/stability.md` completa con hallazgos reales de
        este Pi (zram ya activo con la config default, earlyoom no estaba instalado,
        watchdog de hardware presente pero sin activar en systemd) y comandos exactos.

Al terminar cada fase se prueba en la Pi antes de seguir -- el detalle de cómo probar
cada una vive en el plan de Fase 1 (histórico de la conversación) y se repite al cierre
de cada fase según se va implementando.
