package httpapi

import (
	"encoding/json"
	"net/http"

	"homelab-dashboard/backend/internal/auth"
)

type authHandlers struct {
	passwords *auth.PasswordStore
	store     *auth.Store
	limiter   *auth.RateLimiter
}

type loginRequest struct {
	Password string `json:"password"`
}

func (h *authHandlers) login(w http.ResponseWriter, r *http.Request) {
	ip := auth.ClientIP(r)
	if !h.limiter.Allow(ip) {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "demasiados intentos, espera un minuto"})
		return
	}

	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cuerpo invalido"})
		return
	}

	if !h.passwords.Verify(req.Password) {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "credenciales invalidas"})
		return
	}

	cookieValue, err := h.store.Create()
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "error interno"})
		return
	}
	h.store.SetCookie(w, cookieValue)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

type changePasswordRequest struct {
	CurrentPassword string `json:"current_password"`
	NewPassword     string `json:"new_password"`
}

// changePassword deja actualizar la contrasena de admin desde la propia UI: la nueva
// contrasena se hashea aqui mismo con bcrypt y nunca sale del backend en texto plano
// ni como hash -- ver auth.PasswordStore.Change. Reusa el rate limiter de /api/login
// para que tampoco sirva como via de fuerza bruta contra la contrasena actual.
func (h *authHandlers) changePassword(w http.ResponseWriter, r *http.Request) {
	ip := auth.ClientIP(r)
	if !h.limiter.Allow(ip) {
		writeJSON(w, http.StatusTooManyRequests, map[string]string{"error": "demasiados intentos, espera un minuto"})
		return
	}

	var req changePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "cuerpo invalido"})
		return
	}

	if err := h.passwords.Change(req.CurrentPassword, req.NewPassword); err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *authHandlers) logout(w http.ResponseWriter, r *http.Request) {
	if cookie, err := r.Cookie(auth.CookieName); err == nil {
		h.store.Revoke(cookie.Value)
	}
	h.store.ClearCookie(w)
	writeJSON(w, http.StatusOK, map[string]bool{"ok": true})
}

func (h *authHandlers) session(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(auth.CookieName)
	authenticated := err == nil && h.store.Validate(cookie.Value)
	writeJSON(w, http.StatusOK, map[string]bool{"authenticated": authenticated})
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}
