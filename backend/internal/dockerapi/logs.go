package dockerapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/moby/moby/client"
)

// ansiEscape reconoce secuencias de control ANSI/terminal (limpiar pantalla, mover
// cursor, colores, cambiar titulo de ventana, etc). Muchas imagenes (ej. servidores de
// juegos pensados para correr en una terminal interactiva, como terraria-server) las
// meten en su output -- sin limpiarlas, el visor de logs se llena de basura ilegible y
// el parseo de los adapters (internal/adapters) se rompe.
var ansiEscape = regexp.MustCompile(`\x1b(\[[0-9;?]*[a-zA-Z]|\][^\x07\x1b]*(\x07|\x1b\\))`)

func stripANSI(s string) string {
	return ansiEscape.ReplaceAllString(s, "")
}

// FollowLogs sigue los logs de un contenedor y llama a onLine por cada chunk de output
// ya limpio de codigos ANSI (o cadena vacia si el chunk completo era solo control). Es
// la base compartida tanto del stream SSE de logs para el frontend (StreamLogs) como del
// Manager de adapters (internal/adapters), que necesita las lineas ya separadas del
// transporte HTTP para alimentar Instance.OnLogLine.
//
// since (timestamp unix como string, o "") limita a lineas posteriores a ese momento --
// lo usa el Manager para reconectar sin repetir lo ya leido.
func FollowLogs(ctx context.Context, cli *client.Client, containerID string, tty bool, tail, since string, onLine func(string)) error {
	res, err := cli.ContainerLogs(ctx, containerID, client.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Tail:       tail,
		Since:      since,
	})
	if err != nil {
		return err
	}
	defer res.Close()

	w := lineCallbackWriter(onLine)

	if tty {
		// Con TTY el stream ya viene como un unico flujo (stdout), sin multiplexar.
		_, err = io.Copy(w, res)
	} else {
		// Sin TTY, Docker multiplexa stdout/stderr con un header de 8 bytes por frame.
		_, err = stdcopy.StdCopy(w, w, res)
	}
	if err != nil && ctx.Err() != nil {
		return nil // el contexto se cancelo (cliente cerro la conexion, o el manager para) -- no es un error real
	}
	return err
}

type lineCallbackWriter func(string)

func (f lineCallbackWriter) Write(p []byte) (int, error) {
	clean := stripANSI(string(p))
	if clean != "" {
		f(clean)
	}
	return len(p), nil
}

// StreamLogs escribe los logs en vivo de un contenedor como eventos SSE directo al
// ResponseWriter, hasta que el contexto de la peticion se cancele (el cliente cierra la
// conexion). No pasa por el hub compartido de metricas -- cada peticion de logs abre su
// propio stream independiente, solo mientras la vista de detalle de ese contenedor esta
// montada en el frontend (ver CLAUDE.md: diseno de SSE, maximo 2 streams simultaneos).
func StreamLogs(ctx context.Context, cli *client.Client, w http.ResponseWriter, containerID string, tty bool, tail string) error {
	flusher, ok := w.(http.Flusher)
	if !ok {
		return fmt.Errorf("streaming no soportado")
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	return FollowLogs(ctx, cli, containerID, tty, tail, "", func(line string) {
		payload, err := json.Marshal(line)
		if err != nil {
			return
		}
		fmt.Fprintf(w, "event: log\ndata: %s\n\n", payload)
		flusher.Flush()
	})
}
