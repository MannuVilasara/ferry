package loadbalancer

import (
	"ferry/internal/helper"
	"ferry/internal/logger"
	"net/http"
	"time"
)


func (s *ServerPool) HealthCheck(cfg *helper.HealthCheck) {
	backends := s.GetBackends()

	for _, b := range backends {
		go func(backend *Backend) {

			url := backend.URL.String() + cfg.Path

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				backend.SetAlive(false)
				logger.Info("Backend %s is down: %v", backend.URL, err)
				return
			}

			client := &http.Client{
				Timeout: 2 * time.Second,
			}

			resp, err := client.Do(req)
			if err != nil {
				backend.SetAlive(false)
				logger.Info("Backend %s is down: %v", backend.URL, err)
				return
			} 

			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				backend.SetAlive(true)
				logger.Info("Backend %s is up", backend.URL)
			} else {
				backend.SetAlive(false)
				logger.Info("Backend %s is down: %v", backend.URL, resp.StatusCode)
			}

		}(b)
	}
}