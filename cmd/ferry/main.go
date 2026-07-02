package main

import (
	"ferry/internal/cli"
	"ferry/internal/daemon"
	"ferry/internal/helper"
	"ferry/internal/loadbalancer"
	"ferry/internal/logger"
	"ferry/internal/metrics"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	cli.CheckPermissions()
	shouldExit, configPath := cli.HandleFlags()
	if shouldExit {
		return
	}

	cfg, err := helper.LoadConfig(configPath)
	if err != nil {
		logger.Fatal("Failed to load config: %v", err)
	}

	var strategy loadbalancer.Strategy

	switch cfg.LoadBalancer.Algorithm {
	case "roundrobin":
		strategy = &loadbalancer.RoundRobin{}
	case "leastconn":
		strategy = &loadbalancer.LeastConnection{}
	case "weightedrr":
		strategy = &loadbalancer.WeightedRoundRobin{}
	case "iphashing":
		strategy = &loadbalancer.IPHashing{}
	default:
		strategy = &loadbalancer.RoundRobin{}
	}

	pool := loadbalancer.NewServerPool(strategy)

	var initialBackends []*loadbalancer.Backend
	for _, backend := range cfg.Backends {
		URL, err := url.Parse(backend.Url)
		if err != nil {
			logger.Fatal("Invalid URL: %v", err)
		}

		wgt := backend.Weight
		if wgt <= 0 {
			wgt = 1
		}

		proxy := httputil.NewSingleHostReverseProxy(URL)
		lbbackend := &loadbalancer.Backend{
			Name:    backend.Name,
			URL:     URL,
			Proxy:   proxy,
			IsAlive: true,
			Weight:  wgt,
		}
		initialBackends = append(initialBackends, lbbackend)
		logger.Info("Added backend: %s at %s", backend.Name, backend.Url)
	}
	pool.SetBackends(initialBackends)

	cleanupPID := daemon.ManagePID()
	defer cleanupPID()

	daemon.StartHealthChecker(pool, &cfg.HealthCheck)
	daemon.StartHotReloader(pool, configPath)
	daemon.StartUnixSocketServer(pool)

	go metrics.StartMetricsServer()

	if err = http.ListenAndServe(cfg.Server.Listen, pool); err != nil {
		logger.Fatal("Failed to start server: %v", err)
	}
}
