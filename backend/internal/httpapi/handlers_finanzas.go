package httpapi

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
	"time"
)

// maxFinanzasBody acota lo que el navegador puede mandar a task-hub (los formularios pesan bytes).
const maxFinanzasBody = 64 << 10

// maxReciboBody: el PDF de un recibo de nomina que se sube desde la app (pesan ~70 KB; task-hub
// acepta hasta 5 MB).
const maxReciboBody = 6 << 20

// finanzasRoutes es la lista blanca de lo que la pestaña Dinero puede pedirle a task-hub. Cualquier
// otra ruta o metodo se corta aqui, aunque task-hub la tuviera.
var finanzasRoutes = []struct {
	method string
	path   *regexp.Regexp
}{
	{http.MethodGet, regexp.MustCompile(`^(resumen|reparto|movimientos|nominas|gastos/periodo|cashy|cashy/config|hoy|proximos)$`)},
	{http.MethodPost, regexp.MustCompile(`^(gastos|pagos-recibidos|pagos-tarjeta|saldos-iniciales|cajas|planes|ingresos-extra|cashy|calibrar|simulador|nominas/recibo|tarjetas|entradas)$`)},
	{http.MethodPost, regexp.MustCompile(`^(cajas/\d+/movimientos|tarjetas/[a-z0-9_-]+/ajustar)$`)},
	{http.MethodPut, regexp.MustCompile(`^(tarjetas/[a-z0-9_-]+|cajas/\d+|cashy/config)$`)},
	{http.MethodDelete, regexp.MustCompile(`^((cajas|movimientos|planes)/\d+|cashy)$`)},
}

// finanzasHandlers reenvia /api/finanzas/* a la API de task-hub (dueña del dinero, con su propia
// base) por la red interna taskhub-finanzas. El dashboard no guarda nada: solo agrega el token de
// servicio, que el navegador nunca ve.
type finanzasHandlers struct {
	token string
	proxy *httputil.ReverseProxy
}

func newFinanzasHandlers(baseURL, token string) *finanzasHandlers {
	base, err := url.Parse(strings.TrimRight(baseURL, "/"))
	if err != nil || base.Host == "" {
		base = &url.URL{Scheme: "http", Host: "invalid"}
	}
	h := &finanzasHandlers{token: token}
	h.proxy = &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			path := strings.TrimPrefix(pr.In.URL.Path, "/api/finanzas/")
			pr.Out.URL.Scheme = base.Scheme
			pr.Out.URL.Host = base.Host
			pr.Out.URL.Path = base.Path + "/api/v1/" + path
			pr.Out.URL.RawPath = ""
			pr.Out.Host = base.Host
			pr.Out.Header.Del("Cookie")
			pr.Out.Header.Del("X-Requested-With")
			pr.Out.Header.Set("Authorization", "Bearer "+h.token)
		},
		Transport: &http.Transport{ResponseHeaderTimeout: 10 * time.Second},
		ModifyResponse: func(res *http.Response) error {
			res.Header.Del("Set-Cookie")
			return nil
		},
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			writeJSON(w, http.StatusBadGateway, map[string]string{"error": "task-hub no responde: ¿está encendido?"})
		},
	}
	return h
}

func (h *finanzasHandlers) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if h.token == "" {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "falta FINANZAS_TOKEN en .env.secrets"})
		return
	}
	path := strings.TrimPrefix(r.URL.Path, "/api/finanzas/")
	allowed := false
	for _, rt := range finanzasRoutes {
		if rt.method == r.Method && rt.path.MatchString(path) {
			allowed = true
			break
		}
	}
	if !allowed {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "ruta no permitida"})
		return
	}
	limite := int64(maxFinanzasBody)
	if path == "nominas/recibo" {
		limite = maxReciboBody
	}
	r.Body = http.MaxBytesReader(w, r.Body, limite)
	h.proxy.ServeHTTP(w, r)
}
