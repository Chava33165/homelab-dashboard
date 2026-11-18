# hostctl

Helper mínimo que corre **directo en el host** (no en Docker) vía systemd socket
activation, para las dos operaciones privilegiadas que el contenedor del dashboard
deliberadamente no puede hacer: `reboot`/`poweroff` (vía `systemctl`, autorizado por una
regla polkit dedicada que habla con logind por D-Bus) y `vcgencmd get_throttled` (vía
`/dev/vcio` con permisos de grupo). Ver el diseño completo y el análisis de riesgos en
`CLAUDE.md` ("hostctl: el helper de host") y `docs/security.md`.

Protocolo: JSON-lines sobre un unix socket, 4 comandos fijos (`ping`, `get_throttled`,
`reboot`, `poweroff`), autenticados con un token compartido. Nunca hay `exec.Command`
con input del cliente.

## Instalación (manual, en el host -- no se automatiza desde Docker)

Correr todo esto en el Pi, con `sudo`. Da igual el orden exacto salvo donde se indica.

### 1. Usuarios y grupos

```bash
sudo groupadd --system homelab-hostctl   # comparten: el socket y el token
sudo groupadd --system vcio               # si no existe ya
sudo useradd --system --no-create-home --shell /usr/sbin/nologin \
  -G homelab-hostctl,vcio hostctl
```

### 2. Regla udev para /dev/vcio

```bash
sudo cp udev/99-vcio.rules /etc/udev/rules.d/
sudo udevadm control --reload-rules
sudo udevadm trigger /dev/vcio
ls -la /dev/vcio   # debe mostrar el grupo "vcio"
```

### 3. Regla polkit

```bash
sudo mkdir -p /etc/polkit-1/rules.d
sudo cp polkit/50-homelab-hostctl.rules /etc/polkit-1/rules.d/
sudo systemctl restart polkit
```

### 4. Token compartido con el dashboard

```bash
sudo mkdir -p /etc/homelab-hostctl
sudo chmod 750 /etc/homelab-hostctl
openssl rand -hex 32 | sudo tee /etc/homelab-hostctl/token > /dev/null
sudo chown root:homelab-hostctl /etc/homelab-hostctl/token
sudo chmod 640 /etc/homelab-hostctl/token
```

Copiar ese mismo valor a `HOSTCTL_TOKEN` en el `.env.secrets` del dashboard (el
contenedor lo necesita para autenticar sus peticiones -- ver más abajo).

### 5. Binario

Compilarlo (no hay Go en el host, se hace vía Docker -- ver `make hostctl-build` en la
raíz del repo, ya deja el binario en `hostctl/hostctl-bin`):

```bash
cd ~/Servicios/homelab-dashboard
make hostctl-build
sudo install -m 755 hostctl/hostctl-bin /usr/local/bin/homelab-hostctl
```

### 6. Unidades de systemd

```bash
sudo cp hostctl/systemd/hostctl.socket hostctl/systemd/hostctl.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now hostctl.socket
```

(`hostctl.service` NO se habilita directo -- arranca solo, activado por el socket, la
primera vez que algo se conecta a `/run/homelab-hostctl.sock`.)

### 7. Probar el socket a mano antes de conectarlo al dashboard

```bash
sudo apt install -y socat   # si no está
TOKEN=$(sudo cat /etc/homelab-hostctl/token)
echo "{\"cmd\":\"ping\",\"token\":\"$TOKEN\"}" | socat - UNIX-CONNECT:/run/homelab-hostctl.sock
# debe responder: {"ok":true}

echo "{\"cmd\":\"get_throttled\",\"token\":\"$TOKEN\"}" | socat - UNIX-CONNECT:/run/homelab-hostctl.sock
# debe responder algo como: {"ok":true,"data":{"raw":"0x0"}}

systemctl status hostctl.socket hostctl.service --no-pager
journalctl -u hostctl -n 20 --no-pager
```

### 8. Conectar el dashboard

En `docker-compose.yml`, servicio `dashboard`, descomentar:
```yaml
volumes:
  - /run/homelab-hostctl.sock:/run/homelab-hostctl.sock
group_add:
  - "${HOSTCTL_GID:?}"
```

`HOSTCTL_GID` en `.env` = `getent group homelab-hostctl | cut -d: -f3`. El token va en
`.env.secrets` como `HOSTCTL_TOKEN=<el mismo valor del paso 4>` -- el dashboard lo monta
vía `HOSTCTL_TOKEN_FILE`, no como variable de entorno directa, para no dejarlo en
`docker inspect`.

## Desinstalar

```bash
sudo systemctl disable --now hostctl.socket hostctl.service
sudo rm /etc/systemd/system/hostctl.{socket,service}
sudo rm /etc/polkit-1/rules.d/50-homelab-hostctl.rules
sudo rm /etc/udev/rules.d/99-vcio.rules
sudo rm -rf /etc/homelab-hostctl
sudo rm /usr/local/bin/homelab-hostctl
sudo userdel hostctl
sudo groupdel homelab-hostctl vcio   # vcio solo si nada mas lo usa
sudo systemctl daemon-reload
```
