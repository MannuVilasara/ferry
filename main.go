package main

import (
	"ferry/internal/helper"
	"ferry/internal/loadbalancer"
	"ferry/internal/logger"
	"flag"
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	checkCfg := flag.Bool("c", false, "Check if the Config is valid.")
	checkReload := flag.Bool("r", false, "Reload the config.")

	flag.Parse()

	cfg, err := helper.LoadConfig("config.yaml")
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	if *checkCfg {
		fmt.Printf("Config is valid\n")
		os.Exit(0)
	}

	if *checkReload {
		pidBytes, err := os.ReadFile("/tmp/ferry.pid")
		if err != nil {
			fmt.Printf("Ferry doesn't Seem to be running. Couldn't find /tmp/ferry.pid\n")
			os.Exit(1)
		}

		var pid int

		fmt.Sscanf(string(pidBytes), "%d", &pid)

		ps, err := os.FindProcess(pid)
		if err != nil {
			logger.Fatal("Ferry doesn't seem to be running, send SIGHUP failed: %v", err)
		}

		err = ps.Signal(syscall.SIGHUP)
		if err != nil {
			logger.Fatal("Failed to send signal: %v", err)
		}

		fmt.Printf("Config reloaded successfully\n")
		os.Exit(0)

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

	if cfg.HealthCheck.Enabled {
		go func() {
			t := time.NewTicker(cfg.HealthCheck.Interval)
			defer t.Stop()
			for range t.C {
				pool.HealthCheck(&cfg.HealthCheck)
			}
		}()
	}

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

	pid := os.Getpid()
	logger.Info("Starting Load Balancer -> PID: %d", pid)

	if err = os.WriteFile("/tmp/ferry.pid", []byte(fmt.Sprintf("%d", pid)), 0644); err != nil {
		logger.Fatal("Failed to write PID file: %v", err)
	}

	defer func() {
		if err := os.Remove("/tmp/ferry.pid"); err != nil {
			logger.Error("Failed to remove PID file: %v", err)
		}
	}()

	if err = http.ListenAndServe(cfg.Server.Listen, pool); err != nil {
		logger.Fatal("Failed to start server: %v", err)
	}
}
