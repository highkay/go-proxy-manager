package config

import (
	"os"
	"time"

	"go-proxy-manager/internal/model"
	"gopkg.in/yaml.v3"
)

type AppConfig struct {
	Port        int    `yaml:"port"`
	LogLevel    string `yaml:"log_level"`
	ThreadCount int    `yaml:"thread_count"`
}

type ValidationConfig struct {
	TargetURLs []string      `yaml:"target_urls"`
	Timeout    time.Duration `yaml:"timeout"`
	Interval   time.Duration `yaml:"interval"`
}

type Config struct {
	App        AppConfig        `yaml:"app"`
	Validation ValidationConfig `yaml:"validation"`
	Sources    []model.Source   `yaml:"sources"`
}

func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var cfg Config
	decoder := yaml.NewDecoder(file)
	if err := decoder.Decode(&cfg); err != nil {
		return nil, err
	}

	// Set defaults if necessary
	if cfg.App.Port == 0 {
		cfg.App.Port = 8080
	}
	if cfg.App.ThreadCount == 0 {
		cfg.App.ThreadCount = 50
	}
	if cfg.Validation.Timeout == 0 {
		cfg.Validation.Timeout = 10 * time.Second
	}

	return &cfg, nil
}
