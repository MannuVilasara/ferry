package loadbalancer

import (
	"log"
	"net/http"
	"net/http/httputil"
	"sync/atomic"
)

type ServerPool struct {
	proxies []*httputil.ReverseProxy 
	current uint64                   
}


func NewServerPool() *ServerPool {
	return &ServerPool{
		proxies: make([]*httputil.ReverseProxy, 0),
		current: 0,
	}
}

func (s *ServerPool) AddProxy(proxy *httputil.ReverseProxy) {
	s.proxies = append(s.proxies, proxy)
}

func (s *ServerPool) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	index := uint64(atomic.AddUint64(&s.current, 1)) % uint64(len(s.proxies))

	proxy := s.proxies[index]

	log.Printf("Proxying request to %d", index)
	proxy.ServeHTTP(w,r)
}
