// Comando hostctl: corre directo en el host (no en Docker) via systemd socket
// activation. Expone 4 comandos privilegiados (ping, get_throttled, reboot, poweroff)
// que el contenedor del dashboard no puede hacer por si mismo. Ver hostctl/README.md
// para instalacion y CLAUDE.md para el diseno completo y el analisis de riesgos.
package main

import (
	"bufio"
	"crypto/subtle"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

const (
	tokenPath    = "/etc/homelab-hostctl/token"
	connDeadline = 5 * time.Second
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)
	log.SetPrefix("hostctl: ")

	token, err := loadToken(tokenPath)
	if err != nil {
		log.Fatalf("no se pudo leer el token (%s): %v", tokenPath, err)
	}

	ln, err := systemdListener()
	if err != nil {
		log.Fatalf("socket: %v", err)
	}
	log.Printf("escuchando (activacion por socket de systemd)")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("accept: %v", err)
			continue
		}
		go handleConn(conn, token)
	}
}

func loadToken(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	tok := strings.TrimSpace(string(data))
	if tok == "" {
		return "", fmt.Errorf("archivo de token vacio")
	}
	return tok, nil
}

// systemdListener recibe el socket ya creado por systemd via activacion por socket
// (LISTEN_FDS). No se usa ninguna libreria externa (ej. coreos/go-systemd) a proposito
// -- es un solo fd, siempre en el descriptor 3, la stdlib alcanza de sobra.
func systemdListener() (net.Listener, error) {
	pid, _ := strconv.Atoi(os.Getenv("LISTEN_PID"))
	nfds, _ := strconv.Atoi(os.Getenv("LISTEN_FDS"))
	if pid != os.Getpid() || nfds < 1 {
		return nil, fmt.Errorf(
			"no se recibio activacion por socket de systemd (LISTEN_PID=%q LISTEN_FDS=%q) -- correr via hostctl.socket, no directo",
			os.Getenv("LISTEN_PID"), os.Getenv("LISTEN_FDS"),
		)
	}
	f := os.NewFile(uintptr(3), "hostctl.socket")
	return net.FileListener(f)
}

func handleConn(conn net.Conn, token string) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(connDeadline))

	scanner := bufio.NewScanner(conn)
	if !scanner.Scan() {
		return
	}

	var req request
	if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
		writeResp(conn, errResponse("json invalido"))
		return
	}

	if subtle.ConstantTimeCompare([]byte(req.Token), []byte(token)) != 1 {
		log.Printf("token invalido desde %s (cmd=%s)", conn.RemoteAddr(), req.Cmd)
		writeResp(conn, errResponse("token invalido"))
		return
	}

	log.Printf("comando aceptado: %s", req.Cmd)

	switch req.Cmd {
	case "ping":
		writeResp(conn, okResponse(nil))
	case "get_throttled":
		handleThrottled(conn)
	case "reboot":
		// La respuesta se manda ANTES de ejecutar: reboot corta la conexion (y el
		// proceso) casi de inmediato, no hay garantia de que el cliente alcance a
		// leer nada despues.
		writeResp(conn, okResponse(nil))
		runSystemctl("reboot")
	case "poweroff":
		writeResp(conn, okResponse(nil))
		runSystemctl("poweroff")
	default:
		writeResp(conn, errResponse("comando desconocido"))
	}
}

func handleThrottled(conn net.Conn) {
	out, err := exec.Command("vcgencmd", "get_throttled").Output()
	if err != nil {
		writeResp(conn, errResponse("vcgencmd: "+err.Error()))
		return
	}
	raw := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(string(out)), "throttled="))
	writeResp(conn, okResponse(map[string]string{"raw": raw}))
}

// runSystemctl NUNCA recibe el string de accion desde el cliente -- action siempre es
// un literal fijo ("reboot" o "poweroff") desde el switch de arriba, nunca req.Cmd
// directamente. systemctl reboot/poweroff habla con logind por D-Bus, autorizado por
// la regla polkit dedicada al usuario hostctl (ver polkit/50-homelab-hostctl.rules) --
// mismo mecanismo que D-Bus crudo, sin tener que implementar el protocolo a mano.
func runSystemctl(action string) {
	cmd := exec.Command("systemctl", action)
	if err := cmd.Run(); err != nil {
		log.Printf("systemctl %s: %v", action, err)
	}
}

func writeResp(conn net.Conn, resp response) {
	if err := json.NewEncoder(conn).Encode(resp); err != nil {
		log.Printf("write response: %v", err)
	}
}
