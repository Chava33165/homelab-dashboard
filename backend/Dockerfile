# Fase 4: multi-stage completo. El contexto de build es la raiz del repo (no backend/)
# porque esta etapa necesita frontend/ tambien -- ver "context: ." en docker-compose.yml.

# Etapa 1: build del frontend (Svelte + Vite), estaticos + fuentes autohospedadas.
FROM node:22-alpine AS frontend-build
WORKDIR /frontend
COPY frontend/package.json frontend/package-lock.json ./
RUN npm ci
COPY frontend/ ./
RUN npm run build

# Etapa 2: build del backend (Go), con el frontend ya compilado embebido via go:embed.
FROM golang:1.26-alpine AS build
WORKDIR /src
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend/ .
COPY --from=frontend-build /frontend/dist ./internal/web/dist
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/dashboard ./cmd/dashboard
# /data es donde vive el hash de contrasena mutable en runtime (ver auth.PasswordStore)
# -- se crea aqui, con el dueno ya correcto, porque la imagen final (distroless) no
# tiene shell para hacer mkdir/chown. Docker copia el contenido+permisos de este
# directorio dentro del volumen nombrado "dashboard-data" la primera vez que se monta
# encima de /data (comportamiento estandar de un volumen vacio sobre una ruta con
# contenido en la imagen), asi el proceso (UID 10001, sin root) puede escribir ahi.
RUN mkdir -p /appdata && chown 10001:10001 /appdata

# Etapa 3: runtime minimo con CA certs (necesarios para las llamadas HTTPS del speedtest).
FROM gcr.io/distroless/static-debian12
COPY --from=build /out/dashboard /dashboard
COPY --from=build --chown=10001:10001 /appdata /data
ENTRYPOINT ["/dashboard"]
