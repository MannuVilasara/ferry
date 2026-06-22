package main

import (
	"ferry/internal/helper"
	"ferry/internal/loadbalancer"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	log.Printf("Load Balancer PID: %d\n", os.Getpid())

	cfg, err := helper.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}


	pool := loadbalancer.NewServerPool(&loadbalancer.RoundRobin{})

	var initialBackends []*loadbalancer.Backend

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
		initialBackends = append(initialBackends, lbbackend)

		log.Printf("Added backend: %s at %s", backend.Name, backend.Url)
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

	go func ()  {
		for {
			<-signalChannel
			log.Println("Received SIGHUP! Reloading config...")

			newCfg, err := helper.LoadConfig("config.yaml")
			if err != nil {
				log.Printf("Failed to reload config: %v", err)
				continue
			}

			var newBackend []*loadbalancer.Backend

			for _, backend := range newCfg.Backends {
				URL, err := url.Parse(backend.Url)
				if err != nil {
					log.Printf("Failed to parse URL: %v", err)
					continue
				}
				
				proxy := httputil.NewSingleHostReverseProxy(URL)
				
				lbbackend := &loadbalancer.Backend{
					URL: URL,
					Proxy: proxy,
					IsAlive: true,
				}
				newBackend = append(newBackend, lbbackend)

				log.Printf("Added backend: %s at %s", backend.Name, backend.Url)
				
			}

			pool.SetBackends(newBackend)
			log.Println("Config reloaded successfully")
		}
	}()
	
	if err = http.ListenAndServe(cfg.Server.Listen, pool); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

}
