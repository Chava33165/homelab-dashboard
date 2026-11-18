# Seguridad -- modelo de amenazas

## Qué puede hacer un atacante que compromete el contenedor `dashboard`

El escenario base: un RCE (ejecución remota de código) en el backend Go, por lo que
sea -- una dependencia con una vulnerabilidad, un bug propio. El contenedor ya corre
sin root (`user: 10001:10001`), `cap_drop: [ALL]`, `read_only: true`, sin acceso
directo a `docker.sock`. ¿Qué le queda disponible a un atacante en ese punto?

**Contra Docker** (vía `docker-socket-proxy`): solo lo que el proxy deja pasar --
listar/inspeccionar contenedores, leer logs, pedir stats, y start/stop/restart. **No
puede**: `exec` dentro de un contenedor, `build`/`commit` (que permitirían fabricar una
imagen con `--privileged` o montar `/`), tocar volúmenes, Swarm, secrets, ni nada fuera
de esa lista corta (ver `deploy/docker-socket-proxy.env` para el listado exacto). El
peor caso realista: puede tirar contenedores existentes (DoS a sí mismo/otros
proyectos del Pi), no escapar al host vía Docker.

**Contra `hostctl`**: puede abrir el socket (está en el grupo `homelab-hostctl` vía
`group_add`), pero necesita también el token compartido (`HOSTCTL_TOKEN_FILE`,
montado read-only) para que cualquier comando sea aceptado. Con el token en la mano
(que un RCE en el mismo contenedor SÍ podría leer, ya que está montado ahí), el
abanico se reduce a 4 comandos fijos sin argumentos libres: `ping`, `get_throttled`,
`reboot`, `poweroff`. El peor caso: reiniciar/apagar la Pi (DoS físicamente
recuperable, sin pérdida de datos -- todo lo que importa vive en volúmenes, no en
contenedores efímeros) y leer el estado de throttling (fuga de información de bajo
impacto). `hostctl` en sí corre como usuario dedicado sin privilegios, con
`NoNewPrivileges`, `ProtectSystem=strict`, y solo puede pedir reboot/poweroff vía la
regla polkit específica para ese usuario -- no puede hacer nada más aunque quisiera.

**Contra el resto del Pi vía `pid: host`**: el contenedor puede *ver* la lista de
procesos del host (nombres, PIDs, línea de comandos básica), porque así es como se
leen las métricas de disco/red reales (ver CLAUDE.md, "gotcha" de `/proc/self/...`).
No puede *tocar* nada de eso -- sin `CAP_SYS_PTRACE`, no puede leer
`/proc/<pid>/environ` de procesos de otro usuario, no puede enviarles señales.

**Contra el filesystem del host**: `/proc`, `/sys` y `/` están montados **read-only**.
No hay ninguna ruta de escritura hacia el host desde este contenedor salvo `/tmp`
(tmpfs, efímero, vive y muere con el contenedor).

## Qué puede hacer alguien con solo la URL, sin pasar por el login de la app

Nada -- todas las rutas salvo `/healthz`, `/api/login`, `/api/logout`, `/api/session`
y los estáticos (HTML/JS/CSS/fuentes) exigen sesión válida (`store.RequireAuth`). El
login tiene rate limiting (5 intentos/minuto por IP) y usa bcrypt. La cookie de sesión
es `HttpOnly` (no la lee JS, mitiga XSS robando la sesión) y `SameSite=Strict`
(mitiga CSRF de origen cruzado).

## TCP-connect vs ICMP real para medir latencia de red

Se mide con un `net.DialTimeout("tcp", "1.1.1.1:443", ...)` cronometrado en vez de un
ping ICMP real. Motivo: ICMP real requeriría `cap_add: NET_RAW` en el contenedor (más
superficie) o tocar `net.ipv4.ping_group_range` en el host (config del sistema para
algo que no lo amerita). El connect TCP mide un proxy razonable de "hay salida a
internet y qué tan rápida" sin ampliar nada -- unos pocos ms más que un ICMP real por
el handshake, aceptable para un dashboard, no para diagnóstico de red serio.

## Qué pasa si alguien roba el link/acceso a Tailscale

Es la superficie de ataque real más relevante para este proyecto: cualquiera con
acceso al tailnet (o a quien se le comparta el nodo, ver el proyecto de Terraria en
este mismo repo padre para el mecanismo de compartir) y la contraseña del dashboard
tiene control completo descrito arriba. Por eso: contraseña fuerte (no la de prueba
`test-password-fase2` usada durante el desarrollo -- cambiarla con
`make hash-password PASSWORD=...` antes de considerar esto "en producción"), y por eso
Cloudflare Access (cuando se active la vía pública) es una capa **adicional**, no un
sustituto del login de la app.
