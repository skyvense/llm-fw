package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// ServerConfig defines server configuration
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// OllamaConfig defines Ollama server configuration
type OllamaConfig struct {
	URL string `yaml:"url"`
}

// StorageConfig defines storage configuration
type StorageConfig struct {
	Type string `yaml:"type"`
	Path string `yaml:"path"`
}

// Config defines application configuration
type Config struct {
	Server  ServerConfig  `yaml:"server"`
	Ollama  OllamaConfig  `yaml:"ollama"`
	Storage StorageConfig `yaml:"storage"`
}

// LoadConfig loads configuration from file
func LoadConfig(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %v", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %v", err)
	}

	return &config, nil
}
