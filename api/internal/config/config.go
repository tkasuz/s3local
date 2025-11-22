package config

import (
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Notifications []NotificationRule `json:"notifications" yaml:"notifications"`
}

func LoadConfig() (*Config, error) {
	cfg := &Config{}

	// 1. Load file if exists
	if fileExists("config.yaml") {
		if err := loadYAML("config.yaml", cfg); err != nil {
			return nil, err
		}
	}

	return cfg, nil
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func loadYAML(path string, v any) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(data, v)
}
