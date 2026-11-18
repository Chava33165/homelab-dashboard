package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"strings"
	"sync"
	"time"
)

const (
	CookieName = "homelab_session"
	sessionTTL = 7 * 24 * time.Hour
)

// Store guarda las sesiones vivas en memoria del proceso -- no sobreviven un reinicio
// del contenedor, lo cual es intencional (ver CLAUDE.md: "sin base de datos").
type Store struct {
	secret       []byte
	mu           sync.Mutex
	sessions     map[string]time.Time // sessionID -> expira en
	secureCookie bool
}

func NewStore(secret string, secureCookie bool) *Store {
	s := &Store{
		secret:       []byte(secret),
		sessions:     make(map[string]time.Time),
		secureCookie: secureCookie,
	}
	go s.janitor()
	return s
}

func (s *Store) janitor() {
	ticker := time.NewTicker(10 * time.Minute)
	for range ticker.C {
		now := time.Now()
		s.mu.Lock()
		for id, exp := range s.sessions {
			if now.After(exp) {
				delete(s.sessions, id)
			}
		}
		s.mu.Unlock()
	}
}

func (s *Store) sign(id string) string {
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(id))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

// Create genera una nueva sesion y devuelve el valor listo para meter en el cookie
// (sessionID + firma HMAC, separados por un punto -- asi una peticion con un cookie
// corrupto/forjado se rechaza por la firma sin siquiera tocar el mapa de sesiones).
func (s *Store) Create() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	id := base64.RawURLEncoding.EncodeToString(raw)
	exp := time.Now().Add(sessionTTL)

	s.mu.Lock()
	s.sessions[id] = exp
	s.mu.Unlock()

	return id + "." + s.sign(id), nil
}

func (s *Store) Validate(cookieValue string) bool {
	id, sig, ok := splitCookie(cookieValue)
	if !ok {
		return false
	}
	expected := s.sign(id)
	if subtle.ConstantTimeCompare([]byte(sig), []byte(expected)) != 1 {
		return false
	}

	s.mu.Lock()
	exp, exists := s.sessions[id]
	s.mu.Unlock()

	return exists && time.Now().Before(exp)
}

func (s *Store) Revoke(cookieValue string) {
	id, _, ok := splitCookie(cookieValue)
	if !ok {
		return
	}
	s.mu.Lock()
	delete(s.sessions, id)
	s.mu.Unlock()
}

func splitCookie(v string) (id, sig string, ok bool) {
	i := strings.LastIndexByte(v, '.')
	if i < 0 {
		return "", "", false
	}
	return v[:i], v[i+1:], true
}

func (s *Store) SetCookie(w http.ResponseWriter, value string) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    value,
		Path:     "/",
		HttpOnly: true,
		Secure:   s.secureCookie,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   int(sessionTTL.Seconds()),
	})
}

func (s *Store) ClearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   s.secureCookie,
		SameSite: http.SameSiteStrictMode,
		MaxAge:   -1,
	})
}
