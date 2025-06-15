package config

import (
	"errors"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Name            string   `yaml:"name"`
	Env             []string `yaml:"env"`
	ProtectedRoutes []string `yaml:"protected_routes"`
}

func LoadConfig() (*Config, error) {
	f, err := os.ReadFile("config/dev.yaml")
	if err != nil {
		return nil, errors.New("failed to read config file from root config folder")
	}

	cfg := &Config{}
	err = yaml.Unmarshal(f, &cfg)
	if err != nil {
		return nil, errors.New("failed to unmarshal config file")
	}

	return cfg, nil
}
