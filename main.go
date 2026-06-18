package main

import (
	"ferry/internal/helper"
	"ferry/internal/loadbalancer"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
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
		pool.AddProxy(proxy)

		log.Printf("Added backend: %s at %s", backend.Name, backend.Url)
	}

	if err = http.ListenAndServe(cfg.Server.Listen, pool); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
