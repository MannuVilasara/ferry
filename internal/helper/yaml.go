package helper

import (
	"ferry/internal/logger"
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Server struct {
	Listen string `yaml:"listen"`
}

type LoadBalancer struct {
	Algorithm string `yaml:"algorithm"`
}

type Backend struct {
	Name   string `yaml:"name"`
	Url    string `yaml:"url"`
	Weight int    `yaml:"weight"`
}

type HealthCheck struct {
	Enabled  bool          `yaml:"enabled"`
	Interval time.Duration `yaml:"interval"`
	Path     string        `yaml:"path"`
}

type Logging struct {
	Level string `yaml:"level"`
}

type Metrics struct {
	Enabled bool   `yaml:"enabled"`
	Path    string `yaml:"path"`
}

type Config struct {
	Server       Server       `yaml:"server"`
	LoadBalancer LoadBalancer `yaml:"load_balancer"`
	Backends     []Backend    `yaml:"backends"`
	HealthCheck  HealthCheck  `yaml:"health_check"`
	Logging      Logging      `yaml:"logging"`
	Metrics      Metrics      `yaml:"metrics"`
}

func LoadConfig(path string) (*Config, error) {
	config := &Config{}

	file, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	if err = yaml.Unmarshal(file, config); err != nil {
		return nil, err
	}

	if err := config.Validate(); err != nil {
		return nil, err
	}
	return config, nil
}

func (c *Config) Validate() error {

	SUPPORTED_ALGOS := map[string]struct{}{
		"round_robin": {}, 
		"least_connections": {},
		"weighted_round_robin": {},
	}

	if c.Server.Listen == "" {
		return fmt.Errorf("server.listen shouldn't be empty. Please provide the port number")
	}

	if len(c.Backends) == 0 {
		return fmt.Errorf("backends cannot be empty. Please provide the backends")
	}

	if c.LoadBalancer.Algorithm == "" {
		if len(c.Backends) == 1 {
			c.LoadBalancer.Algorithm = "round_robin"
		} else {
			logger.Warn("load_balancer.algorithm not provided. fallbacking to default round_robin")
			c.LoadBalancer.Algorithm = "round_robin"
		}
	}

	if _, ok := SUPPORTED_ALGOS[c.LoadBalancer.Algorithm]; !ok {
		return fmt.Errorf("load_balancer.algorithm %s is not supported. Supported algorithms are %v", c.LoadBalancer.Algorithm, SUPPORTED_ALGOS)
	}

	if c.HealthCheck.Path == "" {
		logger.Warn("healthcheck.path not provided. fallbacking to default /health")
		c.HealthCheck.Path = "/health"
	}

	if c.HealthCheck.Interval == 0 {
		logger.Warn("healthcheck.interval not provided. fallbacking to default 10s")
		c.HealthCheck.Interval = 10 * time.Second
	}

	if c.HealthCheck.Interval < 1*time.Second {
		return fmt.Errorf("healthcheck.interval should be greater than 1s")
	}

	if c.Logging.Level == "" {
		logger.Warn("logging.level not provided. fallbacking to default info")
		c.Logging.Level = "info"
	}

	if c.Metrics.Path == "" {
		logger.Warn("metrics.path not provided. fallbacking to default /metrics")
		c.Metrics.Path = "/metrics"
	}
	
	return nil
}
