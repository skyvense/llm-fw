package storage

import (
	"llm-fw/api"
	"llm-fw/interfaces"
)

// StorageAdapter 实现了 Storage 接口
type StorageAdapter struct {
	storage interfaces.Storage
}

// NewStorageAdapter 创建一个新的存储适配器
func NewStorageAdapter(storage interfaces.Storage) *StorageAdapter {
	return &StorageAdapter{
		storage: storage,
	}
}

// SaveRequest 保存请求
func (a *StorageAdapter) SaveRequest(req *api.Request) error {
	return a.storage.SaveRequest(req)
}

// GetRequests 获取指定用户的所有请求
func (a *StorageAdapter) GetRequests(userID string) ([]*api.Request, error) {
	return a.storage.GetRequests(userID)
}

// GetAllRequests 获取所有请求
func (a *StorageAdapter) GetAllRequests() ([]*api.Request, error) {
	return a.storage.GetAllRequests()
}

// GetRequestByID 根据ID获取请求
func (a *StorageAdapter) GetRequestByID(requestID string) (*api.Request, error) {
	return a.storage.GetRequestByID(requestID)
}

// DeleteRequest 删除请求
func (a *StorageAdapter) DeleteRequest(requestID string) error {
	return a.storage.DeleteRequest(requestID)
}

// NewHistoryManager 创建一个新的历史记录管理器
func (s *StorageAdapter) NewHistoryManager(size int) interfaces.HistoryManager {
	return NewHistoryManager(size)
}
