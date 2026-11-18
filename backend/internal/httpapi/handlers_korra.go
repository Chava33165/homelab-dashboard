package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"io/fs"
	"net/http"
	"os"
)

// maxKorraStats acota lo que se lee del archivo: el resumen real pesa unos KB.
const maxKorraStats = 1 << 20

// korraHandlers sirve el resumen de uso que genera el asistente KORRA. El dashboard
// no calcula nada ni toca la base del bot: solo lee el archivo montado en solo lectura.
type korraHandlers struct {
	statsFile string
}

func (h *korraHandlers) stats(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open(h.statsFile)
	if errors.Is(err, fs.ErrNotExist) {
		writeJSON(w, http.StatusNotFound, map[string]string{"error": "KORRA todavía no ha generado estadísticas"})
		return
	}
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "no se pudo leer el resumen de KORRA"})
		return
	}
	defer f.Close()

	data, err := io.ReadAll(io.LimitReader(f, maxKorraStats+1))
	if err != nil || len(data) > maxKorraStats || !json.Valid(data) {
		writeJSON(w, http.StatusBadGateway, map[string]string{"error": "el resumen de KORRA está dañado"})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	_, _ = w.Write(data)
}
