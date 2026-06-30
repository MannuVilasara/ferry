package tests

import (
	"testing"
	"ferry/internal/loadbalancer"
)

func TestRoundRobin(t *testing.T) {
	// 1. Setup Fake Backends
	backends := getTestBackends()

	// 2. Initialize Strategy
	lb := &loadbalancer.RoundRobin{}

	// 3. Test the Routing
	
	// Because RoundRobin starts by adding 1 to the atomic counter (0 -> 1)
	// The first index evaluated is 1 % 3 = 1 (Backend2)
	first := lb.NextBackend(backends)
	if first == nil || first.Name != "Backend2" {
		t.Errorf("Expected Backend2, got %v", first)
	}

	// Next is index 2 (Backend3), but it's DEAD!
	// So the algorithm should instantly skip it and return index 3 % 3 = 0 (Backend1)
	second := lb.NextBackend(backends)
	if second == nil || second.Name != "Backend1" {
		t.Errorf("Expected Backend1 (since Backend3 is dead), got %v", second)
	}

	// Next is index 4 % 3 = 1 (Backend2 again)
	third := lb.NextBackend(backends)
	if third == nil || third.Name != "Backend2" {
		t.Errorf("Expected Backend2 again, got %v", third)
	}
}
