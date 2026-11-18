package dockerapi

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"
)

type Stats struct {
	CPUPercent float64 `json:"cpu_percent"`
	MemUsage   uint64  `json:"mem_usage_bytes"`
	MemLimit   uint64  `json:"mem_limit_bytes"`
	MemPercent float64 `json:"mem_percent"`
}

// GetStats pide un sample de stats con IncludePreviousSample:true -- el daemon hace
// el segundo muestreo internamente (con ~1s de espera) y devuelve current+previous en
// una sola respuesta, sin que nosotros tengamos que samplear dos veces a mano.
func GetStats(ctx context.Context, cli *client.Client, containerID string) (Stats, error) {
	res, err := cli.ContainerStats(ctx, containerID, client.ContainerStatsOptions{IncludePreviousSample: true})
	if err != nil {
		return Stats{}, err
	}
	defer res.Body.Close()

	var s container.StatsResponse
	if err := json.NewDecoder(res.Body).Decode(&s); err != nil {
		return Stats{}, err
	}
	return calcStats(s), nil
}

// ContainerWithStats combina el listado con sus stats, pedidas en paralelo (pedirlas en
// serie tardaria ~1s POR contenedor, ya que cada GetStats espera un segundo internamente).
type ContainerWithStats struct {
	ContainerSummary
	Stats *Stats `json:"stats,omitempty"` // nil si no se pudieron obtener (ej. contenedor detenido)
}

func ListWithStats(ctx context.Context, cli *client.Client) ([]ContainerWithStats, error) {
	summaries, err := ListContainers(ctx, cli)
	if err != nil {
		return nil, err
	}

	out := make([]ContainerWithStats, len(summaries))
	var wg sync.WaitGroup
	for i, c := range summaries {
		out[i] = ContainerWithStats{ContainerSummary: c}
		if c.State != "running" {
			continue // un contenedor detenido no tiene stats que pedir
		}
		wg.Add(1)
		go func(i int, id string) {
			defer wg.Done()
			if s, err := GetStats(ctx, cli, id); err == nil {
				out[i].Stats = &s
			}
		}(i, c.ID)
	}
	wg.Wait()
	return out, nil
}

func calcStats(s container.StatsResponse) Stats {
	var cpuPercent float64
	cpuDelta := float64(s.CPUStats.CPUUsage.TotalUsage) - float64(s.PreCPUStats.CPUUsage.TotalUsage)
	sysDelta := float64(s.CPUStats.SystemUsage) - float64(s.PreCPUStats.SystemUsage)
	if sysDelta > 0 && cpuDelta > 0 {
		onlineCPUs := float64(s.CPUStats.OnlineCPUs)
		if onlineCPUs == 0 {
			onlineCPUs = float64(len(s.CPUStats.CPUUsage.PercpuUsage))
		}
		if onlineCPUs == 0 {
			onlineCPUs = 1
		}
		cpuPercent = (cpuDelta / sysDelta) * onlineCPUs * 100
	}

	// Igual que "docker stats": se resta el page cache reclamable de memoria usada,
	// para no contar como "presion" algo que el kernel libera bajo demanda.
	// cgroup v2 expone "inactive_file"; cgroup v1 expone "cache".
	memUsage := s.MemoryStats.Usage
	if v, ok := s.MemoryStats.Stats["inactive_file"]; ok && v < memUsage {
		memUsage -= v
	} else if v, ok := s.MemoryStats.Stats["cache"]; ok && v < memUsage {
		memUsage -= v
	}

	var memPercent float64
	if s.MemoryStats.Limit > 0 {
		memPercent = float64(memUsage) / float64(s.MemoryStats.Limit) * 100
	}

	return Stats{
		CPUPercent: cpuPercent,
		MemUsage:   memUsage,
		MemLimit:   s.MemoryStats.Limit,
		MemPercent: memPercent,
	}
}
