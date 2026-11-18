package metrics

import (
	"time"

	"homelab-dashboard/backend/internal/hostctl"
	"homelab-dashboard/backend/internal/sse"
)

// HostProcPath es el /proc del host montado read-only (ver docker-compose.yml).
const HostProcPath = "/host/proc"

const (
	tickInterval = 2 * time.Second
	procPath     = HostProcPath
	sysPath      = "/host/sys"
	hostRoot     = "/host/root"
	pingTarget   = "1.1.1.1:443"
	pingTimeout  = 1500 * time.Millisecond
)

type NetSnapshot struct {
	RxBytesPerSec map[string]float64 `json:"rx_bytes_per_sec"`
	TxBytesPerSec map[string]float64 `json:"tx_bytes_per_sec"`
	LatencyMs     float64            `json:"latency_ms"`
	Reachable     bool               `json:"reachable"`
}

type Snapshot struct {
	Timestamp  time.Time      `json:"timestamp"`
	CPUPercent float64        `json:"cpu_percent"`
	Mem        MemInfo        `json:"mem"`
	Disks      []DiskUsage    `json:"disks"`
	TempC      float64        `json:"temp_c"`
	Throttle   ThrottleStatus `json:"throttle"`
	Net        NetSnapshot    `json:"net"`
	UptimeSec  float64        `json:"uptime_sec"`
}

// Sampler arma un snapshot de metricas del host cada tickInterval y lo publica en el
// hub SSE. Lee todo de /proc, /sys y disco montados read-only -- nunca via el proxy de
// Docker, sin privilegios especiales (excepto el throttle, que pasa por hostctl).
type Sampler struct {
	hub        *sse.Hub
	hostctlCli *hostctl.Client

	prevCPU  cpuSample
	prevNet  map[string]NetIfaceSample
	prevTick time.Time
}

func NewSampler(hub *sse.Hub, hc *hostctl.Client) *Sampler {
	return &Sampler{hub: hub, hostctlCli: hc, prevNet: map[string]NetIfaceSample{}}
}

func (s *Sampler) Run() {
	s.prevCPU, _ = readCPUSample(procPath)
	s.prevNet, _ = ReadNetSamples(procPath)
	s.prevTick = time.Now()

	ticker := time.NewTicker(tickInterval)
	defer ticker.Stop()
	for range ticker.C {
		s.hub.Publish("metrics", s.sample())
	}
}

func (s *Sampler) sample() Snapshot {
	now := time.Now()
	elapsed := now.Sub(s.prevTick).Seconds()
	if elapsed <= 0 {
		elapsed = tickInterval.Seconds()
	}

	curCPU, _ := readCPUSample(procPath)
	cpuPct := CPUPercent(s.prevCPU, curCPU)
	s.prevCPU = curCPU

	mem, _ := ReadMemInfo(procPath, sysPath)
	disks, _ := ReadDiskUsage(procPath, hostRoot)
	tempC, _ := ReadCPUTempC(sysPath)
	uptime, _ := ReadUptime(procPath)
	throttle := ReadThrottleStatus(s.hostctlCli)

	curNet, _ := ReadNetSamples(procPath)
	rxRates := map[string]float64{}
	txRates := map[string]float64{}
	for iface, cur := range curNet {
		if prev, ok := s.prevNet[iface]; ok && cur.RxBytes >= prev.RxBytes && cur.TxBytes >= prev.TxBytes {
			rxRates[iface] = float64(cur.RxBytes-prev.RxBytes) / elapsed
			txRates[iface] = float64(cur.TxBytes-prev.TxBytes) / elapsed
		}
	}
	s.prevNet = curNet
	s.prevTick = now

	latency, reachable := PingLatency(pingTarget, pingTimeout)

	return Snapshot{
		Timestamp:  now,
		CPUPercent: cpuPct,
		Mem:        mem,
		Disks:      disks,
		TempC:      tempC,
		Throttle:   throttle,
		UptimeSec:  uptime,
		Net: NetSnapshot{
			RxBytesPerSec: rxRates,
			TxBytesPerSec: txRates,
			LatencyMs:     float64(latency.Milliseconds()),
			Reachable:     reachable,
		},
	}
}
