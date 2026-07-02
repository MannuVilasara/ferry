package loadbalancer

import (
	"ferry/internal/logger"
	"net/http"
)

type WeightedRoundRobin struct {
}

func (w *WeightedRoundRobin) NextBackend(backends []*Backend, _ *http.Request) *Backend {
	var targetBackend *Backend
	var totalWeight int64

	for _, b := range backends {
		if b.GetAlive() == false {
			continue
		}

		totalWeight += b.Weight
		logger.Debug("Total weight: %d", totalWeight)

		b.CurrentWeight.Add(b.Weight)

		logger.Debug("Current weight of %s: %d", b.Name, b.CurrentWeight.Load())

		if targetBackend == nil || b.CurrentWeight.Load() > targetBackend.CurrentWeight.Load() {
			targetBackend = b
		}
	}

	if targetBackend == nil {
		return nil
	}

	targetBackend.CurrentWeight.Add(-totalWeight)

	return targetBackend
}
