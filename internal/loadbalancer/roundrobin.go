package loadbalancer

import (
	"net/http"
	"sync/atomic"
)

type RoundRobin struct {
	current atomic.Uint64
}

func (r *RoundRobin) NextBackend(backends []*Backend, _ *http.Request) *Backend {
	index := uint64(r.current.Add(1)) % uint64(len(backends))

	attempts := 0

	for !backends[index].GetAlive() && attempts < len(backends) {
		index = uint64(r.current.Add(1)) % uint64(len(backends))
		attempts++
	}

	if attempts == len(backends) {
		return nil
	}

	return backends[index]

}
