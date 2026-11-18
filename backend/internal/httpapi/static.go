package httpapi

import (
	"io/fs"
	"net/http"
	"strings"
)

// ServeSPA sirve los estaticos embebidos del frontend, con fallback a index.html para
// cualquier ruta que no matchee un archivo real -- el router del frontend usa la
// History API (rutas como /containers/xxx que no existen como archivo en dist/).
func ServeSPA(distFS fs.FS) http.Handler {
	fileServer := http.FileServer(http.FS(distFS))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}
		if _, err := fs.Stat(distFS, path); err != nil {
			r.URL.Path = "/"
		}
		fileServer.ServeHTTP(w, r)
	})
}
