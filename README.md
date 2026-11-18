# homelab-dashboard

> *El panel de control de mi homelab: métricas en vivo, contenedores y dinero, en una sola página.*
> *Self-hosted dashboard for my Raspberry Pi 5 homelab (Go backend, Svelte frontend, Docker).*

## Lo que hace

- **Métricas del host en vivo** (CPU, RAM por servicio, temperatura, red, disco) por **Server-Sent Events**.
- **Control de contenedores** (encender, apagar, reiniciar, logs) a través de un proxy del socket de Docker
  que solo permite lo necesario.
- **Seguridad primero**: sesiones con contraseña (bcrypt), *rate limiting*, proxy a otras APIs con **lista blanca
  de rutas**, contenedores sin root y de solo lectura, acceso por Tailscale y Cloudflare Tunnel.
- Pestaña de **Dinero** conectada a mi [API de finanzas](https://github.com/Chava33165/finanzas-api).

La app de iPhone que usa este backend: [homelab-app](https://github.com/Chava33165/homelab-app).

**Stack:** Go (net/http) · Svelte · Docker multi-stage · docker-socket-proxy · Caddy · Tailscale

Las decisiones de arquitectura están en [CLAUDE.md](CLAUDE.md) y [docs/](docs). Instalación en [docs/INSTALACION.md](docs/INSTALACION.md).

---
Hecho por **Salvador Pantoja** · construido con ayuda de IA (Claude) como herramienta, entendiendo cada pieza.
