# Guía de estabilidad (se configura en el host, no en contenedores)

Todo lo de acá se instala/activa directo en el Pi, fuera de Docker. Requiere `sudo`.

## zram -- ya está activo, no hace falta instalarlo

Confirmado en este Pi: `systemd-zram-generator` (paquete `systemd-zram-generator`, ya
instalado) crea `/dev/zram0` automáticamente al arrancar, usando su configuración por
defecto (`/usr/lib/systemd/zram-generator.conf`): **50% de la RAM o 4GiB, lo que sea
menor**, con el algoritmo de compresión por defecto del kernel (`zstd` en este Pi).
Con 4GB de RAM total, eso da los 2GB de zram que ya se ven en `zramctl`.

Verificar que sigue vivo en cualquier momento:
```bash
zramctl
cat /proc/swaps
```

Si algún día se quisiera ajustar (ej. otro algoritmo o tamaño), se hace con un archivo
propio que sobreescribe el default, **sin tocar el paquete**:
```bash
sudo mkdir -p /etc/systemd/zram-generator.conf.d
sudo tee /etc/systemd/zram-generator.conf.d/homelab.conf > /dev/null <<'EOF'
[zram0]
zram-size = ram / 2
compression-algorithm = zstd
EOF
sudo systemctl daemon-reload
sudo systemctl restart systemd-zram-setup@zram0.service
```
No es necesario ahora mismo -- el default ya es razonable para este Pi.

## earlyoom -- NO está instalado, instalar

Confirmado: `dpkg -l | grep earlyoom` no devuelve nada en este Pi. `earlyoom` mata al
proceso que más memoria usa (o el que el usuario marque como preferido) **antes** de que
el sistema llegue a quedarse sin memoria de verdad y el kernel entre en pánico
matando lo que sea (incluido, potencialmente, sshd o dockerd) -- una muerte
controlada de un contenedor específico es mucho mejor que el sistema entero
congelándose.

```bash
sudo apt update
sudo apt install -y earlyoom
```

Configuración recomendada (`/etc/default/earlyoom`) -- evitar que mate a `sshd` o
`dockerd` por accidente, y preferir matar procesos dentro de contenedores antes que
servicios del host:
```bash
sudo tee /etc/default/earlyoom > /dev/null <<'EOF'
EARLYOOM_ARGS="-r 60 -m 10 -s 10 --avoid '(^|/)(sshd|dockerd|containerd|systemd)$' --prefer '(^|/)(dashboard|terraria-server|node)$'"
EOF
sudo systemctl enable --now earlyoom
systemctl status earlyoom --no-pager
```
`-m 10 -s 10`: actúa cuando queda menos de 10% de RAM Y menos de 10% de swap
disponibles (earlyoom en 2026 reacciona a "memoria disponible", que en un Pi con zram
ya cuenta el swap comprimido correctamente).

## Watchdog de hardware -- confirmado presente, activar vía systemd

Confirmado en este Pi: el driver del watchdog de hardware del Broadcom (Pi 5) ya está
cargado (`/dev/watchdog` y `/dev/watchdog0` existen). No hace falta el paquete
`watchdog` clásico -- systemd (desde hace varias versiones, y este Pi corre systemd
257) puede alimentar el watchdog de hardware **él mismo**, sin un daemon aparte. Si
systemd (PID 1) se cuelga o el kernel entra en pánico y nadie "alimenta" el watchdog a
tiempo, el hardware fuerza un reinicio -- una red de seguridad distinta a earlyoom
(earlyoom cubre presión de memoria; el watchdog cubre que el sistema entero se
congele, kernel panic incluido).

```bash
sudo sed -i \
  -e 's/^#RuntimeWatchdogSec=.*/RuntimeWatchdogSec=15s/' \
  -e 's/^#RebootWatchdogSec=.*/RebootWatchdogSec=10min/' \
  /etc/systemd/system.conf
sudo systemctl daemon-reexec
```
Verificar que systemd está alimentando el watchdog:
```bash
sudo systemctl show -p RuntimeWatchdogUSec
```
(debe mostrar `15000000` -- 15s en microsegundos).

## Diagnosticar un cuelgue después de que ya pasó

Si el Pi se trabó y tuviste que reiniciarlo a la fuerza (o el watchdog lo hizo por
vos), esto es lo que hay que revisar, en orden:

**1. ¿Fue el OOM killer?**
```bash
journalctl -b -1 -k | grep -iE "oom-kill|out of memory|killed process"
```
Una línea como `Out of memory: Killed process 1234 (dashboard)` confirma que el
kernel mató algo por falta de memoria. Si `earlyoom` ya estaba instalado y activo,
revisar también sus propios logs (`journalctl -u earlyoom -b -1`) -- si earlyoom actuó
primero, el kernel OOM killer real nunca debería haber tenido que intervenir.

**2. ¿Fue el watchdog el que reinició?**
```bash
journalctl -b -1 -k | tail -50   # ultimas lineas antes de que se cortara el log
```
Si el log se corta abruptamente sin ningún mensaje de shutdown/reboot limpio
("Reached target Reboot", "Stopped target..."), eso es consistente con que el
watchdog forzó el reinicio porque nadie lo alimentó a tiempo (sistema colgado, no un
reinicio pedido).

**3. ¿Qué contenedor se estaba comiendo la RAM justo antes?**

Si tienen `mem_limit` puestos (ver `docs/retrofitting-mem-limits.md`), Docker mismo
deja rastro de si un contenedor llegó a su límite:
```bash
docker inspect <contenedor> --format '{{.State.OOMKilled}}'
```
`true` significa que el cgroup del contenedor llegó a su `mem_limit` y el kernel mató
un proceso ADENTRO de ese contenedor específico -- contenido, no afectó al resto del
sistema. Es exactamente el escenario que los límites de memoria están diseñados para
lograr.

**4. Vista general rápida del boot anterior:**
```bash
journalctl -b -1 -p err   # solo mensajes de error del boot anterior
```
