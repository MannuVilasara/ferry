package loadbalancer

import (
	"ferry/internal/logger"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

type Backend struct {
	Name    string
	URL     *url.URL
	Proxy   *httputil.ReverseProxy
	IsAlive bool
	mu      sync.RWMutex
}

type Strategy interface {
	NextBackend(backends []*Backend) *Backend
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
	strategy Strategy
	mu sync.RWMutex                
}


func NewServerPool(strategy Strategy) *ServerPool {
	return &ServerPool{
		backends: make([]*Backend, 0),
		strategy: strategy,
	}
}

func (s *ServerPool) GetBackends() []*Backend {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.backends
}

func (s *ServerPool) SetBackends(backends []*Backend){
	s.mu.Lock()
	defer s.mu.Unlock()
	s.backends = backends
}

func (s *ServerPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	backend := s.strategy.NextBackend(s.GetBackends())
	if backend == nil {
		http.Error(w, "All backends are down", http.StatusServiceUnavailable)
		return
	}

	proxy := backend.Proxy
	logger.Info("Proxying request to %s", backend.URL)
	proxy.ServeHTTP(w,r)
}
