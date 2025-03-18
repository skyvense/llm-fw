package storage

import (
	"fmt"

	"llm-fw/config"
	"llm-fw/types"
)

// NewStorage 根据配置创建存储实例
func NewStorage(cfg *config.Config) (types.Storage, error) {
	switch cfg.Storage.Type {
	case config.StorageTypeFile:
		return NewFileStorage(cfg.Storage.Path)
	case config.StorageTypeSQLite:
		return NewSQLiteStorage(cfg.Storage.Path)
	default:
		return nil, fmt.Errorf("不支持的存储类型: %s", cfg.Storage.Type)
	}
}
