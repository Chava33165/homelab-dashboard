package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/moby/moby/client"
	"homelab-dashboard/backend/internal/auth"
	"homelab-dashboard/backend/internal/dockerapi"
)

// sambaHandlers deja ver y cambiar la contraseña de Samba (~/Servicios/samba) desde la app.
// La contraseña vive en un archivo montado (secreto/pass del proyecto samba): la imagen
// dockurr/samba lo lee de /run/secrets/pass y la vuelve a aplicar con smbpasswd en cada
// arranque, asi que cambiarla es escribir el archivo y reiniciar el contenedor (el proxy de
// Docker ya permite restart). Mientras el archivo no exista, Samba usa SAMBA_PASS de su
// .env, que el dashboard nunca lee: por eso GET puede responder "sin contraseña".
type sambaHandlers struct {
	cli       *client.Client
	file      string
	container string
	limiter   *auth.RateLimiter
}

const sambaUser = "chava"

type sambaPasswordRequest struct {
	Password string `json:"password"`
}

func (h *sambaHandlers) getPassword(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store")
	data, err := os.ReadFile(h.file)
	if errors.Is(err, fs.ErrNotExist) {
		writeJSON(w, http.StatusOK, map[string]any{"usuario": sambaUser, "password": nil})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "no se pudo leer la contraseña de Samba"})
		return
	}
	log.Printf("audit: contraseña de Samba consultada desde %s", auth.ClientIP(r))
	writeJSON(w, http.StatusOK, map[string]any{"usuario": sambaUser, "password": strings.TrimRight(string(data), "\n")})
}

func (h *sambaHandlers) setPassword(w http.ResponseWriter, r *http.Request) {
	ip := auth.ClientIP(r)
	if !h.limiter.Allow(ip) {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "demasiados intentos, espera un minuto"})
		return
	}

	var req sambaPasswordRequest
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4096)).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cuerpo invalido"})
		return
	}
	if msg := validSambaPassword(req.Password); msg != "" {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": msg})
		return
	}

	if err := writeFileAtomic(h.file, []byte(req.Password+"\n")); err != nil {
		log.Printf("samba: no se pudo escribir %s: %v", h.file, err)
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "no se pudo guardar la contraseña"})
		return
	}
	log.Printf("audit: contraseña de Samba cambiada desde %s; reiniciando %s", ip, h.container)

	// El reinicio tarda unos segundos; si la app se cierra a medias no debe quedar sin aplicar.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(r.Context()), 60*time.Second)
	defer cancel()
	if err := dockerapi.RestartContainer(ctx, h.cli, h.container); err != nil {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "la contraseña se guardó, pero Samba no se pudo reiniciar: " + err.Error()})
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

// validSambaPassword devuelve "" si sirve. smbpasswd la recibe por stdin linea por linea,
// asi que no puede llevar saltos de linea ni otros caracteres de control.
func validSambaPassword(p string) string {
	switch {
	case !utf8.ValidString(p):
		return "la contraseña tiene caracteres invalidos"
	case utf8.RuneCountInString(p) < 8:
		return "la contraseña debe tener al menos 8 caracteres"
	case utf8.RuneCountInString(p) > 128:
		return "la contraseña puede tener como maximo 128 caracteres"
	case strings.IndexFunc(p, unicode.IsControl) >= 0:
		return "la contraseña no puede llevar saltos de linea ni caracteres de control"
	}
	return ""
}

func writeFileAtomic(path string, data []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(path), ".pass-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if err := tmp.Chmod(0o640); err != nil {
		tmp.Close()
		return err
	}
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}
