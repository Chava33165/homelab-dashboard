# Acceso: Tailscale (activo) + Cloudflare Tunnel/Access (preparado, sin activar)

## Vía Tailscale/LAN -- la que está activa

HTTPS real via `tailscale cert` (Let's Encrypt, cadena públicamente confiada) +
Caddy como reverse proxy. Sin CA interna, sin perfiles que instalar en el iPhone.

### Configuración inicial (una sola vez)

```bash
# 1. Permitir que tu usuario use `tailscale cert` sin sudo cada vez:
sudo tailscale set --operator=$USER

# 2. Generar el primer certificado:
mkdir -p /tmp && cd ~/Servicios/homelab-dashboard
sudo mkdir -p /var/lib/homelab-dashboard/tls
tailscale cert \
  --cert-file=/var/lib/homelab-dashboard/tls/homelab.tu-tailnet.ts.net.crt \
  --key-file=/var/lib/homelab-dashboard/tls/homelab.tu-tailnet.ts.net.key \
  homelab.tu-tailnet.ts.net

# 3. Levantar Caddy (ya no tiene profile, arranca con el resto):
docker compose up -d caddy

# 4. Instalar la renovación automática (el cert no se renueva solo):
sudo cp deploy/systemd/tailscale-cert-renew.{service,timer} /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable --now tailscale-cert-renew.timer
```

### El link

```
https://homelab.tu-tailnet.ts.net
```

Funciona desde cualquier dispositivo dentro del tailnet (o con el nodo compartido,
como se hizo con el servidor de Terraria) -- **incluso en la misma LAN de casa**,
Tailscale usa la ruta directa local automáticamente (confirmado en este mismo Pi
durante el proyecto de Terraria: cuando dos dispositivos están en la misma red,
Tailscale detecta la conexión directa y no pierde velocidad pasando por relay). No
hay razón para tener también una URL de LAN "pelada" (IP:puerto sin TLS) -- por eso el
dashboard ya no publica el puerto 8080 directo, todo pasa por Caddy en 443.

### Verificar que el certificado renueva bien

```bash
sudo systemctl start tailscale-cert-renew.service  # correrlo a mano una vez para probar
journalctl -u tailscale-cert-renew -n 20 --no-pager
```

## Vía Cloudflare Tunnel + Access -- preparado, NO activado

El usuario pidió quedarse por ahora solo en LAN + Tailscale. Cuando se quiera activar
el acceso público (mismo dominio que ya usa `task-hub`):

1. Crear un túnel nuevo en el dashboard de Cloudflare Zero Trust, apuntando el
   hostname público a `http://dashboard:8080` (el nombre del servicio interno, no
   "localhost" -- `cloudflared` corre en la red Docker `cloudflare-net`, junto al
   dashboard).
2. Poner el token del túnel en `.env` (`CLOUDFLARE_TUNNEL_TOKEN`).
3. **Antes de exponerlo**, configurar una aplicación de **Cloudflare Access** para ese
   mismo hostname, exigiendo login (email OTP o Google/GitHub) -- decisión ya tomada
   con el usuario: este panel puede reiniciar/apagar la Pi y controlar todos los
   contenedores, así que el login propio de la app NO es la única barrera aceptable
   para la vía pública. Ver CLAUDE.md, sección "Acceso: dos vías".
4. Recién entonces:
   ```bash
   docker compose --profile public up -d cloudflared
   ```

`docker-socket-proxy` nunca comparte red con `cloudflared` (confirmado en
`docker-compose.yml`, redes separadas) -- aunque la vía pública tuviera un problema,
el proxy de Docker queda inalcanzable desde ahí.
