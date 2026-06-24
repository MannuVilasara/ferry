package main

import (
	"ferry/internal/cli"
	"ferry/internal/daemon"
	"ferry/internal/helper"
	"ferry/internal/loadbalancer"
	"ferry/internal/logger"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	cli.CheckPermissions()

	if cli.HandleFlags() {
		return
	}

	cfg, err := helper.LoadConfig("config.yaml")
	if err != nil {
		logger.Fatal("Failed to load config: %v", err)
	}

	pool := loadbalancer.NewServerPool(&loadbalancer.RoundRobin{})

	var initialBackends []*loadbalancer.Backend
	for _, backend := range cfg.Backends {
		URL, err := url.Parse(backend.Url)
		if err != nil {
			logger.Fatal("Invalid URL: %v", err)
		}

		proxy := httputil.NewSingleHostReverseProxy(URL)
		lbbackend := &loadbalancer.Backend{
			URL:     URL,
			Proxy:   proxy,
			IsAlive: true,
		}
		initialBackends = append(initialBackends, lbbackend)
		logger.Info("Added backend: %s at %s", backend.Name, backend.Url)
	}
	pool.SetBackends(initialBackends)

	cleanupPID := daemon.ManagePID()
	defer cleanupPID()

	daemon.StartHealthChecker(pool, &cfg.HealthCheck)
	daemon.StartHotReloader(pool)

	if err = http.ListenAndServe(cfg.Server.Listen, pool); err != nil {
		logger.Fatal("Failed to start server: %v", err)
	}
}
