package auth

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/crypto/bcrypt"
)

// PasswordStore guarda el hash bcrypt del admin. Arranca desde el hash de bootstrap
// (ADMIN_PASSWORD_BCRYPT_HASH, ver .env.secrets) y en su primer uso lo persiste en un
// archivo dentro de un volumen con estado (DataDir/admin_password_hash) -- de ahi en
// adelante ese archivo es la fuente de verdad. Esto permite cambiar la contrasena desde
// la propia UI sin recrear el contenedor ni volver a tocar .env.secrets: el cambio
// queda en memoria al instante y en disco para que sobreviva un reinicio.
type PasswordStore struct {
	path string
	mu   sync.RWMutex
	hash string
}

func NewPasswordStore(dataDir, bootstrapHash string) (*PasswordStore, error) {
	s := &PasswordStore{path: filepath.Join(dataDir, "admin_password_hash")}

	data, err := os.ReadFile(s.path)
	if err == nil {
		s.hash = strings.TrimSpace(string(data))
		if s.hash == "" {
			return nil, fmt.Errorf("%s existe pero esta vacio", s.path)
		}
		return s, nil
	}
	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("leyendo %s: %w", s.path, err)
	}

	if bootstrapHash == "" {
		return nil, fmt.Errorf("no hay contrasena configurada: falta %s y ADMIN_PASSWORD_BCRYPT_HASH", s.path)
	}
	s.hash = bootstrapHash
	if err := s.persist(); err != nil {
		return nil, err
	}
	return s, nil
}

// Verify compara la contrasena en texto plano contra el hash actual.
func (s *PasswordStore) Verify(password string) bool {
	s.mu.RLock()
	hash := s.hash
	s.mu.RUnlock()
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password)) == nil
}

// Change valida la contrasena actual y, si es correcta, reemplaza el hash en memoria
// (efecto inmediato, sin reiniciar el contenedor) y en disco (sobrevive un reinicio).
func (s *PasswordStore) Change(currentPassword, newPassword string) error {
	if len(newPassword) < 8 {
		return fmt.Errorf("la contrasena nueva debe tener al menos 8 caracteres")
	}
	if !s.Verify(currentPassword) {
		return fmt.Errorf("la contrasena actual no es correcta")
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("generando hash: %w", err)
	}

	s.mu.Lock()
	old := s.hash
	s.hash = string(newHash)
	s.mu.Unlock()

	if err := s.persist(); err != nil {
		// Si no se pudo guardar en disco, revertimos el cambio en memoria -- mejor
		// fallar por completo que dejar el proceso con una contrasena que no
		// sobrevive un reinicio (el usuario perderia acceso sin darse cuenta).
		s.mu.Lock()
		s.hash = old
		s.mu.Unlock()
		return fmt.Errorf("no se pudo guardar la nueva contrasena: %w", err)
	}
	return nil
}

func (s *PasswordStore) persist() error {
	s.mu.RLock()
	hash := s.hash
	s.mu.RUnlock()

	if err := os.MkdirAll(filepath.Dir(s.path), 0o700); err != nil {
		return err
	}
	tmp := s.path + ".tmp"
	if err := os.WriteFile(tmp, []byte(hash+"\n"), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, s.path)
}
