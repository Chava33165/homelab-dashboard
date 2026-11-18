#!/usr/bin/env bash
# Renueva el certificado de Tailscale para este Pi y reinicia Caddy para que lo tome.
# tailscale cert NO se renueva solo -- este script lo corre el timer de systemd
# (tailscale-cert-renew.timer, mensual). Necesita correr como root (o con el operador
# de tailscale seteado) para escribir en /var/lib/homelab-dashboard/tls.
set -euo pipefail

DOMAIN="homelab.tu-tailnet.ts.net"
CERT_DIR="/var/lib/homelab-dashboard/tls"
COMPOSE_DIR="/home/usuario/Servicios/homelab-dashboard"

mkdir -p "$CERT_DIR"
tailscale cert --cert-file="$CERT_DIR/$DOMAIN.crt" --key-file="$CERT_DIR/$DOMAIN.key" "$DOMAIN"

cd "$COMPOSE_DIR"
docker compose --profile tls restart caddy
