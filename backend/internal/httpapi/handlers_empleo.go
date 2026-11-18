package httpapi

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// empleoRoutes es la lista blanca de lo que la app (pestaña Jobs) puede pedirle a empleo-api
// (~/Servicios/empleo). Lo demas se corta aqui. La entrevista y editar el perfil siguen solo en la pagina
// Empleo de ~/Servicios/paginas-web.
var empleoRoutes = []struct {
	method string
	path   *regexp.Regexp
}{
	{http.MethodGet, regexp.MustCompile(`^(resumen|vacantes|contactos|perfil|cvs|bandeja)$`)},
	{http.MethodPost, regexp.MustCompile(`^avisar$`)},
	{http.MethodPatch, regexp.MustCompile(`^vacantes/[0-9a-f]{6,16}$`)},
}

const maxEmpleoBody = 16 << 10

// empleoHandlers reenvia /api/empleo/* a empleo-api por la red paginas_web. empleo-api no tiene login
// propio (en la web lo pone Caddy); aqui lo pone RequireAuth del dashboard.
type empleoHandlers struct {
	proxy *httputil.ReverseProxy
}

func newEmpleoHandlers(baseURL string) *empleoHandlers {
	base, err := url.Parse(strings.TrimRight(baseURL, "/"))
	if err != nil || base.Host == "" {
		base = &url.URL{Scheme: "http", Host: "invalid"}
	}
	return &empleoHandlers{proxy: &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			path := strings.TrimPrefix(pr.In.URL.Path, "/api/empleo/")
			pr.Out.URL.Scheme = base.Scheme
			pr.Out.URL.Host = base.Host
			pr.Out.URL.Path = base.Path + "/" + path
			pr.Out.URL.RawPath = ""
			pr.Out.Host = base.Host
			pr.Out.Header.Del("Cookie")
			pr.Out.Header.Del("Authorization")
			// empleo-api exige este encabezado en las escrituras (defensa CSRF de la pagina web).
			pr.Out.Header.Set("X-Requested-With", "empleo")
		},
		Transport: &http.Transport{ResponseHeaderTimeout: 10 * time.Second},
		ModifyResponse: func(res *http.Response) error {
			res.Header.Del("Set-Cookie")
			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "empleo-api no responde: ¿está encendido?"})
		},
	}}
}

func (h *empleoHandlers) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/empleo/")
	for _, rt := range empleoRoutes {
		if rt.method == r.Method && rt.path.MatchString(path) {
			r.Body = http.MaxBytesReader(w, r.Body, maxEmpleoBody)
			h.proxy.ServeHTTP(w, r)
			return
		}
	}
	writeJSON(w, http.StatusNotFound, map[string]string{"error": "ruta no permitida"})
}
