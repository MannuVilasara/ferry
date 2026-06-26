package loadbalancer

type LeastConnection struct{}


func (l *LeastConnection) NextBackend(backends []*Backend) *Backend {
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
	