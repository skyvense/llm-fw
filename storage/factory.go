package storage

import (
	"fmt"

	"llm-fw/config"
	"llm-fw/types"
)

// NewStorage creates a new storage instance based on configuration
func NewStorage(cfg *config.Config) (types.Storage, error) {
	switch cfg.Storage.Type {
	case config.StorageTypeFile:
		return NewFileStorageImpl(cfg.Storage.Path)
	case config.StorageTypeSQLite:
		return NewSQLiteStorage(cfg.Storage.Path)
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", cfg.Storage.Type)
	}
}
