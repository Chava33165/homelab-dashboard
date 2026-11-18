package auth

import "net/http"

// RequireAuth exige un cookie de sesion valido. Se aplica por ruta (no globalmente)
// porque /api/login, /api/logout y /api/session deben quedar accesibles sin sesion.
func (s *Store) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(CookieName)
		if err != nil || !s.Validate(cookie.Value) {
			http.Error(w, `{"error":"no autenticado"}`, http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// RequireCSRFHeader exige un header que solo un fetch/XHR same-origin puede mandar
// (una peticion cross-site "simple", como un <form> de otro origen, no puede agregar
// headers custom sin disparar CORS preflight). Combinado con SameSite=Strict en la
// cookie de sesion, es suficiente para este caso de uso: un solo usuario, sin
// superficie multiusuario ni CORS habilitado hacia otros origenes.
func RequireCSRFHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
			if r.Header.Get("X-Requested-With") != "homelab-dashboard" {
				http.Error(w, `{"error":"falta header requerido"}`, http.StatusForbidden)
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}
