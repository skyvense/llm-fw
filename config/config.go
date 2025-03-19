package config

import (
	"fmt"
	"os"
	"strconv"

	"gopkg.in/yaml.v3"
)

// StorageType 定义存储类型
type StorageType string

const (
	StorageTypeFile   StorageType = "file"
	StorageTypeSQLite StorageType = "sqlite"
)

// Config 表示应用程序配置
type Config struct {
	Server struct {
		Host string `yaml:"host"`
		Port int    `yaml:"port"`
	} `yaml:"server"`
	Ollama struct {
		URL string `yaml:"url"`
	} `yaml:"ollama"`
	Storage struct {
		Type StorageType `yaml:"type"`
		Path string      `yaml:"path"`
	} `yaml:"storage"`
}

// NewConfig 创建新的配置实例
func NewConfig() *Config {
	cfg := &Config{
		Server: struct {
			Host string `yaml:"host"`
			Port int    `yaml:"port"`
		}{
			Host: "0.0.0.0",
			Port: 8080,
		},
		Ollama: struct {
			URL string `yaml:"url"`
		}{
			URL: "http://localhost:11434",
		},
		Storage: struct {
			Type StorageType `yaml:"type"`
			Path string      `yaml:"path"`
		}{
			Type: StorageTypeFile,
			Path: "./data",
		},
	}

	// 从环境变量加载配置
	if host := os.Getenv("SERVER_HOST"); host != "" {
		cfg.Server.Host = host
	}
	if port := os.Getenv("SERVER_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.Server.Port = p
		}
	}
	if url := os.Getenv("OLLAMA_URL"); url != "" {
		cfg.Ollama.URL = url
	}
	if storageType := os.Getenv("STORAGE_TYPE"); storageType != "" {
		cfg.Storage.Type = StorageType(storageType)
	}
	if path := os.Getenv("STORAGE_PATH"); path != "" {
		cfg.Storage.Path = path
	}

	return cfg
}

// LoadConfig 从文件加载配置
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// 规范化存储类型
	switch cfg.Storage.Type {
	case "file":
		cfg.Storage.Type = StorageTypeFile
	case "sqlite":
		cfg.Storage.Type = StorageTypeSQLite
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", cfg.Storage.Type)
	}

	return &cfg, nil
}
