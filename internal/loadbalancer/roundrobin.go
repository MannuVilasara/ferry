package loadbalancer

import (
	"ferry/internal/helper"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
	"time"
)

type Backend struct {
	URL *url.URL
	Proxy *httputil.ReverseProxy
	IsAlive bool
	mu sync.RWMutex
}

func (b *Backend) SetAlive(state bool){
	b.mu.Lock()
	defer b.mu.Unlock()
	b.IsAlive = state
}

func (b *Backend) GetAlive()bool{
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.IsAlive
}

type ServerPool struct {
	backends []*Backend
	current uint64                   
}


func NewServerPool() *ServerPool {
	return &ServerPool{
		backends: make([]*Backend, 0),
		current: 0,
	}
}

func (s *ServerPool) AddBackend(backend *Backend) {
	s.backends = append(s.backends, backend)
}

func (s *ServerPool) HealthCheck(cfg *helper.HealthCheck) {
	for _, b := range s.backends {
		go func(backend *Backend) {

			url := backend.URL.String() + cfg.Path

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				backend.SetAlive(false)
				log.Printf("Backend %s is down: %v", backend.URL, err)
				return
			}

			client := &http.Client{
				Timeout: 2 * time.Second,
			}

			resp, err := client.Do(req)
			if err != nil {
				backend.SetAlive(false)
				log.Printf("Backend %s is down: %v", backend.URL, err)
				return
			} 

			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				backend.SetAlive(true)
				log.Printf("Backend %s is up", backend.URL)
			} else {
				backend.SetAlive(false)
				log.Printf("Backend %s is down: %v", backend.URL, resp.StatusCode)
			}

		}(b)
	}
}

func (s *ServerPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	

	index := uint64(atomic.AddUint64(&s.current, 1)) % uint64(len(s.backends))

	attempts := 0

	for !s.backends[index].GetAlive() && attempts < len(s.backends) {
		index = uint64(atomic.AddUint64(&s.current, 1)) % uint64(len(s.backends))
		attempts++
	}

	if attempts == len(s.backends) {
		http.Error(w, "All backends are down", http.StatusServiceUnavailable)
		return
	}

	proxy := s.backends[index].Proxy

	log.Printf("Proxying request to %s", s.backends[index].URL)
	proxy.ServeHTTP(w,r)
}
