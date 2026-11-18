# mem_limit para los proyectos existentes (task-hub, terraria-server)

**Ya aplicado** (Fase 6) -- esto documenta lo que se hizo, no un plan pendiente.

## Medido en vivo antes de decidir los límites

```bash
docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}"
```

| Contenedor | Uso medido | Límite elegido | Razonamiento |
|---|---|---|---|
| `terraria-server` | ~940MB-1GB con 2 jugadores (medido en la Fase 2 de ese proyecto); 167MB en reposo | **2048m** | Margen para 8 jugadores + mundo más activo (más NPCs/ítems tirados) |
| `task-hub-backend-1` | ~39MB en reposo | **256m** | Backend chico, margen amplio |
| `task-hub-cloudflared-1` | ~32MB en reposo | **128m** | `cloudflared` se mantiene liviano, margen amplio |

## Diffs aplicados

`~/terraria-server/docker-compose.yml`:
```diff
     restart: unless-stopped
     stdin_open: true
     tty: true
+    mem_limit: 2048m
     ports:
```

`~/Servicios/task-hub/docker-compose.yml`:
```diff
     image: ghcr.io/chava33165/task-hub-backend:latest
     restart: unless-stopped
+    mem_limit: 256m
     env_file: .env
```
```diff
   cloudflared:
     image: cloudflare/cloudflared:latest
     restart: unless-stopped
+    mem_limit: 128m
     command: tunnel run
```

Aplicado con `docker compose up -d` en cada proyecto (recrea el contenedor, necesario
para que tome el límite -- un `restart` normal no alcanza si el contenedor ya existía
antes de que el cgroup de memoria estuviera habilitado, ver
`~/Servicios/homelab-dashboard/CLAUDE.md`).

## Verificado a nivel cgroup real, no solo config

```bash
CID=$(docker inspect <contenedor> --format '{{.Id}}')
cat /sys/fs/cgroup/system.slice/docker-${CID}.scope/memory.max
```

Confirmado: `terraria-server` → `2147483648` (2048MB), `task-hub-backend-1` →
`268435456` (256MB), `task-hub-cloudflared-1` → `134217728` (128MB). Los tres
coinciden exactamente con lo configurado -- el kernel los está aplicando de verdad.

## Total de RAM reservada vs. disponible

Sumando todos los `mem_limit` del Pi (homelab-dashboard 256m + docker-socket-proxy 64m
+ caddy 64m + terraria-server 2048m + task-hub-backend 256m + task-hub-cloudflared
128m) da **~2.8GB** de los 4GB totales -- deja margen para el sistema operativo y zram.
Si en el futuro se agrega otro servicio pesado, revisar esta suma antes de asumir que
"total" mem_limit configurado es lo mismo que RAM realmente disponible (los límites son
techos individuales, no una reserva garantizada -- si todos los contenedores llegan a
su techo simultáneamente sí importa que la suma quepa).
