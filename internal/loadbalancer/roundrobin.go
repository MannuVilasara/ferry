package loadbalancer

import (
	"sync/atomic"
)


type RoundRobin struct {
	current uint64
}

func (r *RoundRobin) NextBackend(backends []*Backend) *Backend {
	index := uint64(atomic.AddUint64(&r.current, 1)) % uint64(len(backends))

	attempts := 0

	for !backends[index].GetAlive() && attempts < len(backends) {
		index = uint64(atomic.AddUint64(&r.current, 1)) % uint64(len(backends))
		attempts++
	}

	if attempts == len(backends) {
		return nil
	}

	return backends[index]

}