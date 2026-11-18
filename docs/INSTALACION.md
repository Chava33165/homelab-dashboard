# homelab-dashboard

Dashboard web para monitorear y controlar el homelab (Raspberry Pi 5, 4GB ARM64).
Ver [CLAUDE.md](./CLAUDE.md) para las decisiones de arquitectura completas.

## Estado

Fase 1 de 6 -- estructura y arquitectura. Ver la sección "Fases" en `CLAUDE.md` para el
detalle y el checklist de avance.

## Desarrollo rapido

```bash
cp .env.example .env   # llenar valores reales, .env nunca se versiona
make config             # valida docker-compose.yml
make build               # build de la imagen del dashboard (funcional desde Fase 2)
make up                  # levanta docker-socket-proxy + dashboard
```

No hay Go ni Node instalados en el host -- todo se construye dentro de Docker
(ver Dockerfile multi-stage en `backend/`).
