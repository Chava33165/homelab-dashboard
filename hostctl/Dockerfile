# Este Dockerfile NUNCA se corre como contenedor en produccion -- hostctl vive en el
# host, no en Docker. Es solo un artefacto de compilacion cruzada: no hay Go instalado
# en el host, asi que se compila aqui (arm64, estatico) y se extrae el binario via
# `make hostctl-build` (ver Makefile en la raiz del repo) para instalarlo manualmente.

FROM golang:1.26-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o /hostctl .

FROM scratch
COPY --from=build /hostctl /hostctl
# CMD dummy: este stage nunca se ejecuta de verdad (solo se hace "docker create" +
# "docker cp" para extraer el binario), pero "docker create" exige que haya un
# entrypoint/comando valido en la imagen o falla con "no command specified".
CMD ["/hostctl"]
