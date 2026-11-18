package httpapi

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/moby/moby/client"
	"homelab-dashboard/backend/internal/adapters"
	"homelab-dashboard/backend/internal/auth"
	"homelab-dashboard/backend/internal/dockerapi"
)

type dockerHandlers struct {
	cli     *client.Client
	adapter *adapters.Manager
}

func (h *dockerHandlers) list(w http.ResponseWriter, r *http.Request) {
	list, err := dockerapi.ListWithStats(r.Context(), h.cli)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, h.adapter.Annotate(list))
}

func (h *dockerHandlers) logsStream(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	tty, err := dockerapi.IsTTY(r.Context(), h.cli, id)
	if err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}

	if err := dockerapi.StreamLogs(r.Context(), h.cli, w, id, tty, "200"); err != nil {
		// La conexion ya pudo haber empezado a escribir la respuesta -- no hay un
		// status code limpio que mandar aqui, solo se loguea del lado del servidor.
		return
	}
}

func (h *dockerHandlers) start(w http.ResponseWriter, r *http.Request) {
	h.control(w, r, dockerapi.StartContainer)
}

func (h *dockerHandlers) stop(w http.ResponseWriter, r *http.Request) {
	h.control(w, r, dockerapi.StopContainer)
}

func (h *dockerHandlers) restart(w http.ResponseWriter, r *http.Request) {
	h.control(w, r, dockerapi.RestartContainer)
}

type controlFunc func(ctx context.Context, cli *client.Client, id string) error

func (h *dockerHandlers) control(w http.ResponseWriter, r *http.Request, action controlFunc) {
	id := r.PathValue("id")
	log.Printf("audit: %s pidio control de contenedor %s desde %s", r.Pattern, id, auth.ClientIP(r))
	if err := action(r.Context(), h.cli, id); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *dockerHandlers) serviceOn(w http.ResponseWriter, r *http.Request) {
	h.servicePower(w, r, true)
}

func (h *dockerHandlers) serviceOff(w http.ResponseWriter, r *http.Request) {
	h.servicePower(w, r, false)
}

// servicePower no usa el contexto del request tal cual: si el usuario cierra la
// pestaña a medio apagar, cortar ahi dejaria el servicio a medias. Se desacopla de la
// cancelacion pero con un tope (varios stops de ~10s cada uno en el peor caso).
func (h *dockerHandlers) servicePower(w http.ResponseWriter, r *http.Request, on bool) {
	project := r.PathValue("project")
	verb := "apagar"
	if on {
		verb = "encender"
	}
	log.Printf("audit: %s pidio %s el servicio %s desde %s", r.Pattern, verb, project, auth.ClientIP(r))

	ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), 3*time.Minute)
	defer cancel()

	err := dockerapi.SetServicePower(ctx, h.cli, project, on)
	switch {
	case errors.Is(err, dockerapi.ErrServiceNotFound):
		writeJSON(w, http.StatusNotFound, map[string]string{"error": err.Error()})
	case errors.Is(err, dockerapi.ErrServiceProtected):
		writeJSON(w, http.StatusForbidden, map[string]string{"error": err.Error()})
	case err != nil:
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": err.Error()})
	default:
		writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
	}
}
