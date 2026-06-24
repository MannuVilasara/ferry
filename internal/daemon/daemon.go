package daemon

import (
	"ferry/internal/helper"
	"ferry/internal/loadbalancer"
	"ferry/internal/logger"
	"fmt"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func StartHealthChecker(pool *loadbalancer.ServerPool, cfg *helper.HealthCheck) {
	if !cfg.Enabled {
		return
	}

	go func() {
		t := time.NewTicker(cfg.Interval)
		defer t.Stop()
		for range t.C {
			pool.HealthCheck(cfg)
		}
	}()
}

func StartHotReloader(pool *loadbalancer.ServerPool) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGHUP)

	go func() {
		for {
			<-signalChannel
			logger.Info("Received SIGHUP! Reloading config...")

			newCfg, err := helper.LoadConfig("config.yaml")
			if err != nil {
				logger.Info("Failed to reload config: %v", err)
				continue
			}

			var newBackend []*loadbalancer.Backend
			for _, backend := range newCfg.Backends {
				URL, err := url.Parse(backend.Url)
				if err != nil {
					logger.Info("Failed to parse URL: %v", err)
					continue
				}

				proxy := httputil.NewSingleHostReverseProxy(URL)
				lbbackend := &loadbalancer.Backend{
					URL:     URL,
					Proxy:   proxy,
					IsAlive: true,
				}
				newBackend = append(newBackend, lbbackend)
				logger.Info("Added backend: %s at %s", backend.Name, backend.Url)
			}

			pool.SetBackends(newBackend)
			logger.Info("Config reloaded successfully")
		}
	}()
}

func ManagePID() func() {
	pid := os.Getpid()
	logger.Info("Starting Load Balancer -> PID: %d", pid)

	if err := os.WriteFile("/tmp/ferry.pid", []byte(fmt.Sprintf("%d", pid)), 0644); err != nil {
		logger.Fatal("Failed to write PID file: %v", err)
	}

	return func() {
		if err := os.Remove("/tmp/ferry.pid"); err != nil {
			logger.Error("Failed to remove PID file: %v", err)
		}
	}
}
