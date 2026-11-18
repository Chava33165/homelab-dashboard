package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"homelab-dashboard/backend/internal/adapters"
	_ "homelab-dashboard/backend/internal/adapters/generic"  // se registra via init(), ver adapters/registry.go
	_ "homelab-dashboard/backend/internal/adapters/terraria" // idem
	"homelab-dashboard/backend/internal/auth"
	"homelab-dashboard/backend/internal/config"
	"homelab-dashboard/backend/internal/dockerapi"
	"homelab-dashboard/backend/internal/hostctl"
	"homelab-dashboard/backend/internal/httpapi"
	"homelab-dashboard/backend/internal/metrics"
	"homelab-dashboard/backend/internal/sse"
	"homelab-dashboard/backend/internal/web"
)

func main() {
	healthcheck := flag.Bool("healthcheck", false, "hace un self-check HTTP y sale (usado por el HEALTHCHECK de Docker)")
	flag.Parse()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if *healthcheck {
		os.Exit(runHealthcheck(cfg.ListenAddr))
	}

	passwords, err := auth.NewPasswordStore(cfg.DataDir, cfg.AdminPasswordHash)
	if err != nil {
		log.Fatalf("contrasena de admin: %v", err)
	}

	store := auth.NewStore(cfg.SessionSecret, cfg.CookieSecure)
	limiter := auth.NewRateLimiter()
	hub := sse.NewHub()
	hc := hostctl.NewClient(cfg.HostctlSocket, cfg.HostctlTokenFile)

	sampler := metrics.NewSampler(hub, hc)
	go sampler.Run()

	dockerCli, err := dockerapi.NewClient(cfg.DockerProxyURL)
	if err != nil {
		log.Fatalf("docker client: %v", err)
	}

	bgCtx, bgCancel := context.WithCancel(context.Background())
	defer bgCancel()
	adapterMgr := adapters.NewManager(dockerCli, hub)
	go adapterMgr.Run(bgCtx)

	annotate := func(list []dockerapi.ContainerWithStats) any { return adapterMgr.Annotate(list) }
	go dockerapi.NewSampler(dockerCli, hub, annotate).Run(bgCtx)
	go dockerapi.WatchEvents(bgCtx, dockerCli, hub)

	staticFS, err := web.FS()
	if err != nil {
		log.Fatalf("frontend embebido: %v", err)
	}

	router := httpapi.NewRouter(passwords, store, limiter, hub, dockerCli, adapterMgr, hc, cfg.KorraStatsFile, cfg.FinanzasURL, cfg.FinanzasToken, cfg.EmpleoURL, cfg.SambaPassFile, cfg.SambaContainer, staticFS)
	srv := &http.Server{Addr: cfg.ListenAddr, Handler: router}

	go func() {
		log.Printf("escuchando en %s", cfg.ListenAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("servidor: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
}

func runHealthcheck(listenAddr string) int {
	addr := listenAddr
	if len(addr) > 0 && addr[0] == ':' {
		addr = "127.0.0.1" + addr
	}
	client := http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(fmt.Sprintf("http://%s/healthz", addr))
	if err != nil {
		return 1
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return 1
	}
	return 0
}
