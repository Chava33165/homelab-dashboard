#!/usr/bin/env bash
# Genera el hash bcrypt de una contrasena y lo escapa para .env.secrets.
#
# Por que hace falta el escapado: docker compose interpola "$" tambien dentro de los
# archivos usados con "env_file:" (no solo en docker-compose.yml). Un hash bcrypt
# siempre trae varios "$" (ej. "$2a$14$..."), y sin escapar, compose intenta resolver
# "$2a", "$14", etc. como variables y corrompe el valor. Escapando cada "$" como "$$"
# el valor llega intacto al contenedor. Confirmado empiricamente contra esta version de
# Docker Compose -- ver CLAUDE.md.
set -euo pipefail

if [ -z "${1:-}" ]; then
  echo "uso: $0 <password>" >&2
  exit 1
fi

hash=$(docker run --rm caddy:2-alpine caddy hash-password --plaintext "$1")
escaped=$(printf '%s' "$hash" | sed 's/\$/$$/g')

echo "ADMIN_PASSWORD_BCRYPT_HASH=${escaped}"
echo "" >&2
echo "Copia esa linea completa a .env.secrets (reemplazando la existente)." >&2
