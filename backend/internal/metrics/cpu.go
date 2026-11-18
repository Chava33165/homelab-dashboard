package metrics

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type cpuSample struct {
	idle  uint64
	total uint64
}

// readCPUSample lee la primera linea de /proc/stat ("cpu  user nice system idle ...").
// El indice 3 (despues de quitar la etiqueta "cpu") es el campo idle.
func readCPUSample(procPath string) (cpuSample, error) {
	f, err := os.Open(procPath + "/stat")
	if err != nil {
		return cpuSample{}, err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	if !scanner.Scan() {
		return cpuSample{}, scanner.Err()
	}

	fields := strings.Fields(scanner.Text())
	var sample cpuSample
	for i, field := range fields[1:] {
		v, err := strconv.ParseUint(field, 10, 64)
		if err != nil {
			continue
		}
		sample.total += v
		if i == 3 {
			sample.idle = v
		}
	}
	return sample, nil
}

// CPUPercent calcula el % de uso de CPU entre dos samples consecutivos de /proc/stat.
func CPUPercent(prev, curr cpuSample) float64 {
	totalDelta := float64(curr.total - prev.total)
	idleDelta := float64(curr.idle - prev.idle)
	if totalDelta <= 0 {
		return 0
	}
	return (1 - idleDelta/totalDelta) * 100
}
