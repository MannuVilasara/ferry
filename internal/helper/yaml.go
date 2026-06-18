package helper

import (
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
	return config, nil
}
