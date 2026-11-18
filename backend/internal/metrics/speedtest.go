package metrics

import (
	"context"
	"fmt"

	"github.com/showwin/speedtest-go/speedtest"
)

type SpeedtestResult struct {
	DownloadMbps float64 `json:"download_mbps"`
	UploadMbps   float64 `json:"upload_mbps"`
	LatencyMs    float64 `json:"latency_ms"`
	Server       string  `json:"server"`
}

// RunSpeedtest corre un speedtest bajo demanda (nunca automatico, siempre disparado por
// el usuario) usando una libreria Go pura -- sin shell-out a un binario externo, para
// mantener la imagen final minima.
func RunSpeedtest(ctx context.Context) (SpeedtestResult, error) {
	servers, err := speedtest.FetchServerListContext(ctx)
	if err != nil {
		return SpeedtestResult{}, fmt.Errorf("no se pudo obtener lista de servidores: %w", err)
	}
	targets, err := servers.FindServer(nil)
	if err != nil || len(targets) == 0 {
		return SpeedtestResult{}, fmt.Errorf("no se encontro un servidor de prueba: %w", err)
	}
	server := targets[0]

	if err := server.PingTestContext(ctx, nil); err != nil {
		return SpeedtestResult{}, fmt.Errorf("ping test: %w", err)
	}
	if err := server.DownloadTestContext(ctx); err != nil {
		return SpeedtestResult{}, fmt.Errorf("download test: %w", err)
	}
	if err := server.UploadTestContext(ctx); err != nil {
		return SpeedtestResult{}, fmt.Errorf("upload test: %w", err)
	}

	return SpeedtestResult{
		DownloadMbps: server.DLSpeed.Mbps(),
		UploadMbps:   server.ULSpeed.Mbps(),
		LatencyMs:    float64(server.Latency.Milliseconds()),
		Server:       server.Name,
	}, nil
}
