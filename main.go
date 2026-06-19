package main

import (
	"ferry/internal/helper"
	"ferry/internal/loadbalancer"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

func main() {
	cfg, err := helper.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	pool := loadbalancer.NewServerPool()

	for _, backend := range cfg.Backends {
		
		URL, err := url.Parse(backend.Url)
		if err != nil {
			log.Fatalf("Invalid URL: %v", err)
		}

		proxy := httputil.NewSingleHostReverseProxy(URL)

		lbbackend := &loadbalancer.Backend{
			URL: URL,
			Proxy: proxy,
			IsAlive: true,
		}
		pool.AddBackend(lbbackend)

		log.Printf("Added backend: %s at %s", backend.Name, backend.Url)
	}

	if cfg.HealthCheck.Enabled {
		go func() {
			t := time.NewTicker(cfg.HealthCheck.Interval)
			defer t.Stop()
			for range t.C {
				pool.HealthCheck(&cfg.HealthCheck)
			}
		}()
	}

	if err = http.ListenAndServe(cfg.Server.Listen, pool); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
