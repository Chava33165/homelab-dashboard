package metrics

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

type MemInfo struct {
	TotalKB     uint64  `json:"total_kb"`
	AvailableKB uint64  `json:"available_kb"`
	UsedKB      uint64  `json:"used_kb"`
	UsedPercent float64 `json:"used_percent"`
	SwapTotalKB uint64  `json:"swap_total_kb"`
	SwapUsedKB  uint64  `json:"swap_used_kb"`
	ZramActive  bool    `json:"zram_active"`
}

func ReadMemInfo(procPath, sysPath string) (MemInfo, error) {
	f, err := os.Open(procPath + "/meminfo")
	if err != nil {
		return MemInfo{}, err
	}
	defer f.Close()

	vals := map[string]uint64{}
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		parts := strings.Fields(scanner.Text())
		if len(parts) < 2 {
			continue
		}
		key := strings.TrimSuffix(parts[0], ":")
		v, err := strconv.ParseUint(parts[1], 10, 64)
		if err != nil {
			continue
		}
		vals[key] = v // todo viene en kB
	}

	info := MemInfo{
		TotalKB:     vals["MemTotal"],
		AvailableKB: vals["MemAvailable"],
		SwapTotalKB: vals["SwapTotal"],
		ZramActive:  zramActive(sysPath),
	}
	info.UsedKB = info.TotalKB - info.AvailableKB
	if info.TotalKB > 0 {
		info.UsedPercent = float64(info.UsedKB) / float64(info.TotalKB) * 100
	}
	info.SwapUsedKB = info.SwapTotalKB - vals["SwapFree"]

	return info, nil
}

// zramActive detecta si hay un dispositivo zram activo -- no lo esta todavia en este
// Pi (se configura en Fase 6 segun docs/stability.md), asi que hoy devuelve false.
func zramActive(sysPath string) bool {
	_, err := os.Stat(sysPath + "/block/zram0")
	return err == nil
}
