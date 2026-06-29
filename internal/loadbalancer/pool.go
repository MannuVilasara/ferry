package loadbalancer

import (
	"ferry/internal/logger"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"sync/atomic"
)

type Backend struct {
	// default stuff
	Name    string
	URL     *url.URL
	Proxy   *httputil.ReverseProxy
	IsAlive bool

	// for least connection strategy
	connections atomic.Int64

	// for weighted round robin strategy
	Weight        int64
	CurrentWeight atomic.Int64

	// mutex for locking
	mu sync.RWMutex
}

type Strategy interface {
	NextBackend(backends []*Backend) *Backend
}

func (b *Backend) SetAlive(state bool) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.IsAlive = state
}

func (b *Backend) GetAlive() bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.IsAlive
}

func (b *Backend) IncrementCount() {
	b.connections.Add(1)
}

func (b *Backend) DecrementCount() {
	b.connections.Add(-1)
}

func (b *Backend) GetCount() int64 {
	return b.connections.Load()
}

type ServerPool struct {
	backends []*Backend
	strategy Strategy
	mu       sync.RWMutex
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

func (s *ServerPool) SetBackends(backends []*Backend) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.backends = backends
}

func (s *ServerPool) SetStrategy(strategy Strategy) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.strategy = strategy
}

func (s *ServerPool) GetStrategy() Strategy {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.strategy
}

func (s *ServerPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	backend := s.GetStrategy().NextBackend(s.GetBackends())
	if backend == nil {
		http.Error(w, "All backends are down", http.StatusServiceUnavailable)
		return
	}

	if _, ok := s.GetStrategy().(*LeastConnection); ok {
		backend.IncrementCount()
		defer backend.DecrementCount()
	}

	proxy := backend.Proxy
	logger.Info("Proxying request to %s", backend.URL)
	proxy.ServeHTTP(w, r)
}
