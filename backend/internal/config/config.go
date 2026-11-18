package config

import (
	"fmt"
	"os"
)

type Config struct {
	ListenAddr        string
	DockerProxyURL    string
	HostctlSocket     string
	HostctlTokenFile  string
	AdminPasswordHash string
	DataDir           string
	SessionSecret     string
	CookieSecure      bool
	KorraStatsFile    string
	FinanzasURL       string
	FinanzasToken     string
	EmpleoURL         string
	SambaPassFile     string
	SambaContainer    string
}

func Load() (*Config, error) {
	cfg := &Config{
		ListenAddr:       getEnv("LISTEN_ADDR", ":8080"),
		DockerProxyURL:   os.Getenv("DOCKER_PROXY_URL"),
		HostctlSocket:    getEnv("HOSTCTL_SOCKET", "/run/homelab-hostctl.sock"),
		HostctlTokenFile: os.Getenv("HOSTCTL_TOKEN_FILE"),
		// AdminPasswordHash es solo el hash de bootstrap (ver .env.secrets) para la
		// PRIMERA vez que arranca el contenedor -- una vez que existe
		// DataDir/admin_password_hash, ese archivo manda (ver auth.PasswordStore).
		// Por eso ya no es obligatorio aqui: solo falla si tampoco hay archivo.
		AdminPasswordHash: os.Getenv("ADMIN_PASSWORD_BCRYPT_HASH"),
		DataDir:           getEnv("DATA_DIR", "/data"),
		SessionSecret:     os.Getenv("SESSION_SECRET"),
		CookieSecure:      getEnv("COOKIE_SECURE", "true") == "true",
		// Resumen que escribe el asistente KORRA (~/asistente) cada 15 min; montado ro.
		KorraStatsFile: getEnv("KORRA_STATS_FILE", "/korra/estadisticas.json"),
		// API de dinero (~/task-hub) por la red interna taskhub-finanzas; el token vive en
		// .env.secrets (mismo BEARER_TOKEN que usan los Atajos del iPhone).
		FinanzasURL:   getEnv("FINANZAS_URL", "http://task-hub:8000"),
		FinanzasToken: os.Getenv("FINANZAS_TOKEN"),
		// API de la pagina Empleo (~/Servicios/empleo) por la red paginas_web: pestaña Jobs de la app.
		EmpleoURL: getEnv("EMPLEO_URL", "http://empleo-api:8000"),
		// Contraseña de Samba (~/Servicios/samba/secreto montado en /samba) y el contenedor
		// que hay que reiniciar para aplicarla. Ver httpapi/handlers_samba.go.
		SambaPassFile:  getEnv("SAMBA_PASS_FILE", "/samba/pass"),
		SambaContainer: getEnv("SAMBA_CONTAINER", "samba-samba-1"),
	}

	if len(cfg.SessionSecret) < 16 {
		return nil, fmt.Errorf("SESSION_SECRET debe estar configurado (al menos 16 caracteres)")
	}

	return cfg, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
