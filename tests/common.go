package tests

import "ferry/internal/loadbalancer"

// getTestBackends returns a standard list of mock backends for testing algorithms.
func getTestBackends() []*loadbalancer.Backend {
	backend1 := &loadbalancer.Backend{Name: "Backend1", IsAlive: true, Weight: 4}
	backend2 := &loadbalancer.Backend{Name: "Backend2", IsAlive: true, Weight: 2}
	backend3 := &loadbalancer.Backend{Name: "Backend3", IsAlive: false, Weight: 1} // DEAD!

	return []*loadbalancer.Backend{backend1, backend2, backend3}
}