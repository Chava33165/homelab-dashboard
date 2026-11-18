package metrics

import (
	"bufio"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type NetIfaceSample struct {
	RxBytes uint64
	TxBytes uint64
}

// ReadNetSamples lee /proc/1/net/dev (no /proc/net/dev -- mismo motivo que en disk.go:
// es un symlink a /proc/self/net/dev, que resuelve al namespace de red del propio
// contenedor, no al del host) y devuelve bytes rx/tx acumulados por interfaz (sin
// loopback). El sampler calcula throughput en vivo restando dos samples consecutivos.
func ReadNetSamples(procPath string) (map[string]NetIfaceSample, error) {
	f, err := os.Open(procPath + "/1/net/dev")
	if err != nil {
		return nil, err
	}
	defer f.Close()

	out := map[string]NetIfaceSample{}
	scanner := bufio.NewScanner(f)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		if lineNum <= 2 { // dos lineas de encabezado
			continue
		}
		parts := strings.SplitN(scanner.Text(), ":", 2)
		if len(parts) != 2 {
			continue
		}
		iface := strings.TrimSpace(parts[0])
		if iface == "lo" {
			continue
		}
		fields := strings.Fields(parts[1])
		if len(fields) < 9 {
			continue
		}
		rx, _ := strconv.ParseUint(fields[0], 10, 64)
		tx, _ := strconv.ParseUint(fields[8], 10, 64)
		out[iface] = NetIfaceSample{RxBytes: rx, TxBytes: tx}
	}
	return out, nil
}

// PingLatency mide el tiempo de un connect TCP a addr (ej. "1.1.1.1:443") en vez de un
// ICMP real -- evita pedir cap_add:NET_RAW o tocar sysctls del host. Mide un proxy
// razonable de salud de red, no un ping exacto. Ver CLAUDE.md.
func PingLatency(addr string, timeout time.Duration) (time.Duration, bool) {
	start := time.Now()
	conn, err := net.DialTimeout("tcp", addr, timeout)
	if err != nil {
		return 0, false
	}
	defer conn.Close()
	return time.Since(start), true
}
