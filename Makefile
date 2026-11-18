.PHONY: config build up down logs hostctl-build hash-password

# Valida sintaxis/interpolacion de docker-compose.yml sin necesitar builds funcionales.
config:
	docker compose config

# Funcional a partir de Fase 2 (necesita backend/Dockerfile real).
build:
	docker compose build dashboard

up:
	docker compose up -d docker-socket-proxy dashboard

down:
	docker compose down

logs:
	docker compose logs -f dashboard

# Genera la linea ADMIN_PASSWORD_BCRYPT_HASH ya escapada para .env.secrets.
# Uso: make hash-password PASSWORD=tu-password
hash-password:
	@./scripts/hash-password.sh "$(PASSWORD)"

# Fase 6: compila el binario de hostctl para arm64 via Docker (no hay Go en el host)
# y lo extrae a ./hostctl/hostctl-bin para instalarlo manualmente en el host.
hostctl-build:
	docker build -f hostctl/Dockerfile -t hostctl-build hostctl
	docker create --name hostctl-extract hostctl-build
	docker cp hostctl-extract:/hostctl ./hostctl/hostctl-bin
	docker rm hostctl-extract
