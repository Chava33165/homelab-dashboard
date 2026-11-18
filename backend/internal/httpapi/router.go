package httpapi

import (
	"io/fs"
	"net/http"

	"github.com/moby/moby/client"
	"homelab-dashboard/backend/internal/adapters"
	"homelab-dashboard/backend/internal/auth"
	"homelab-dashboard/backend/internal/hostctl"
	"homelab-dashboard/backend/internal/sse"
)

func NewRouter(passwords *auth.PasswordStore, store *auth.Store, limiter *auth.RateLimiter, hub *sse.Hub, dockerCli *client.Client, adapterMgr *adapters.Manager, hc *hostctl.Client, korraStatsFile, finanzasURL, finanzasToken, empleoURL, sambaPassFile, sambaContainer string, staticFS fs.FS) http.Handler {
	mux := http.NewServeMux()

	ah := &authHandlers{passwords: passwords, store: store, limiter: limiter}
	mh := &metricsHandlers{hub: hub}
	dh := &dockerHandlers{cli: dockerCli, adapter: adapterMgr}
	hh := &hostHandlers{hostctl: hc}
	kh := &korraHandlers{statsFile: korraStatsFile}
	fh := newFinanzasHandlers(finanzasURL, finanzasToken)
	eh := newEmpleoHandlers(empleoURL)
	sh := &sambaHandlers{cli: dockerCli, file: sambaPassFile, container: sambaContainer, limiter: limiter}

	// Sin auth: /healthz (usado por el HEALTHCHECK de Docker) y el propio flujo de login.
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})
	mux.HandleFunc("POST /api/login", ah.login)
	mux.HandleFunc("POST /api/logout", ah.logout)
	mux.HandleFunc("GET /api/session", ah.session)

	// Requieren sesion valida:
	mux.Handle("POST /api/auth/change-password", store.RequireAuth(http.HandlerFunc(ah.changePassword)))
	mux.Handle("GET /api/stream", store.RequireAuth(http.HandlerFunc(mh.stream)))
	mux.Handle("GET /api/memory/processes", store.RequireAuth(http.HandlerFunc(mh.hostProcesses)))
	mux.Handle("POST /api/speedtest", store.RequireAuth(http.HandlerFunc(mh.speedtest)))

	mux.Handle("GET /api/containers", store.RequireAuth(http.HandlerFunc(dh.list)))
	mux.Handle("GET /api/containers/{id}/logs/stream", store.RequireAuth(http.HandlerFunc(dh.logsStream)))
	mux.Handle("POST /api/containers/{id}/start", store.RequireAuth(http.HandlerFunc(dh.start)))
	mux.Handle("POST /api/containers/{id}/stop", store.RequireAuth(http.HandlerFunc(dh.stop)))
	mux.Handle("POST /api/containers/{id}/restart", store.RequireAuth(http.HandlerFunc(dh.restart)))
	mux.Handle("POST /api/services/{project}/start", store.RequireAuth(http.HandlerFunc(dh.serviceOn)))
	mux.Handle("POST /api/services/{project}/stop", store.RequireAuth(http.HandlerFunc(dh.serviceOff)))

	mux.Handle("GET /api/korra/stats", store.RequireAuth(http.HandlerFunc(kh.stats)))
	mux.Handle("/api/finanzas/", store.RequireAuth(fh))
	mux.Handle("/api/empleo/", store.RequireAuth(eh))

	mux.Handle("GET /api/samba/password", store.RequireAuth(http.HandlerFunc(sh.getPassword)))
	mux.Handle("PUT /api/samba/password", store.RequireAuth(http.HandlerFunc(sh.setPassword)))

	mux.Handle("POST /api/host/reboot", store.RequireAuth(http.HandlerFunc(hh.reboot)))
	mux.Handle("POST /api/host/poweroff", store.RequireAuth(http.HandlerFunc(hh.poweroff)))

	mux.Handle("/", ServeSPA(staticFS))

	return SecurityHeaders(auth.RequireCSRFHeader(mux))
}
