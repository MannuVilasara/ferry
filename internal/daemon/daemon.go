package daemon

import (
	"encoding/json"
	"ferry/internal/helper"
	"ferry/internal/loadbalancer"
	"ferry/internal/logger"
	"fmt"
	"net"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

type BackendStatus struct {
	Name    string `json:"name"`
	URL     string `json:"url"`
	IsAlive bool   `json:"is_alive"`
}


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

func StartHotReloader(pool *loadbalancer.ServerPool, configPath string) {
	signalChannel := make(chan os.Signal, 1)
	signal.Notify(signalChannel, syscall.SIGHUP)

	go func() {
		for {
			<-signalChannel
			logger.Info("Received SIGHUP. Reloading config...")
			newCfg, err := helper.LoadConfig(configPath)
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
					Name:    backend.Name,
					URL:     URL,
					Proxy:   proxy,
					IsAlive: true,
				}
				newBackend = append(newBackend, lbbackend)
				logger.Info("Added backend: %s at %s", backend.Name, backend.Url)
			}

			var strategy loadbalancer.Strategy
			switch newCfg.LoadBalancer.Algorithm {
			case "roundrobin":
				strategy = &loadbalancer.RoundRobin{}
			case "leastconn":
				strategy = &loadbalancer.LeastConnection{}
			default:
				strategy = &loadbalancer.RoundRobin{}
			}
			
			pool.SetStrategy(strategy)
			pool.SetBackends(newBackend)
			logger.Info("Config reloaded successfully")
		}
	}()
}

func ManagePID() func() {
	pid := os.Getpid()
	logger.Info("Starting Load Balancer -> PID: %d", pid)

	if err := os.WriteFile("/tmp/ferry.pid", fmt.Appendf(nil, "%d", pid), 0644); err != nil {
		logger.Fatal("Failed to write PID file: %v", err)
	}

	return func() {
		if err := os.Remove("/tmp/ferry.pid"); err != nil {
			logger.Error("Failed to remove PID file: %v", err)
		}
	}
}

func StartUnixSocketServer(pool *loadbalancer.ServerPool) {

	os.Remove("/tmp/ferry.sock")
	sock, err := net.Listen("unix", "/tmp/ferry.sock")
	if err != nil {
		logger.Fatal("Failed to start unix socket: %v", err)
	}
	os.Chmod("/tmp/ferry.sock", 0660)
	
	go func() {
		defer sock.Close()
		for {
		conn, err := sock.Accept()
		if err != nil {
			continue
		}
		go func (c net.Conn, p *loadbalancer.ServerPool)  {
			defer c.Close()

			rawBackends := p.GetBackends()

			backends := make([]BackendStatus, len(rawBackends))

			for i, b := range rawBackends {
				backends[i] = BackendStatus{
					Name: b.Name,
					URL: b.URL.String(),
					IsAlive: b.GetAlive(),
				}
			}

			response := struct {
				ActiveBackends int           `json:"active_backends"`
				TotalBackends  int           `json:"total_backends"`
				Backends       []BackendStatus `json:"backends"`
			}{
				ActiveBackends: 0,
				TotalBackends:  len(backends),
				Backends:       backends,
			}

			for _, b := range backends {
				if b.IsAlive {
					response.ActiveBackends++
				}
			}
			
			json.NewEncoder(c).Encode(response)
			

			
		}(conn, pool)
		}
	}()
}
	
	