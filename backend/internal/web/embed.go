// Package web embebe el frontend compilado (frontend/dist, generado por Vite) dentro
// del binario de Go via //go:embed. El Dockerfile multi-stage construye el frontend en
// una etapa con node:alpine y copia el resultado a internal/web/dist ANTES del build de
// Go -- por eso dist/ no existe en el repo (esta en .gitignore), solo en tiempo de build.
package web

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distFS embed.FS

// FS devuelve el contenido de frontend/dist listo para servir, sin el prefijo "dist/".
func FS() (fs.FS, error) {
	return fs.Sub(distFS, "dist")
}
