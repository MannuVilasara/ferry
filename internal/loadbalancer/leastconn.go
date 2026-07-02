package loadbalancer

import "net/http"

type LeastConnection struct{}

func (l *LeastConnection) NextBackend(backends []*Backend, req *http.Request) *Backend {
	var targetBackend *Backend

	for _, b := range backends {
		if b.GetAlive() {
			if targetBackend == nil || b.GetCount() < targetBackend.GetCount() {
				targetBackend = b
			}
		}
	}
	if targetBackend == nil {
		return nil
	}

	return targetBackend
}
