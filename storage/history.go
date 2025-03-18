package storage

import (
	"container/ring"
	"sort"
	"sync"
	"time"

	"llm-fw/api"
	"llm-fw/interfaces"
)

// HistoryManager 管理请求历史记录
type HistoryManager struct {
	storage interfaces.Storage
	history *ring.Ring
	mu      sync.RWMutex
	size    int
}

// NewHistoryManager 创建一个新的历史记录管理器
func NewHistoryManager(storage interfaces.Storage, size int) *HistoryManager {
	return &HistoryManager{
		storage: storage,
		history: ring.New(size),
		size:    size,
	}
}

// AddEntry 添加一条新的历史记录
func (hm *HistoryManager) AddEntry(entry api.Request) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	entry.Timestamp = time.Now()
	hm.history.Value = entry
	hm.history = hm.history.Next()

	return nil
}

// GetHistory 获取历史记录
func (hm *HistoryManager) GetHistory() ([]*api.Request, error) {
	// 从存储中获取所有请求
	requests, err := hm.storage.GetAllRequests()
	if err != nil {
		return nil, err
	}

	// 按时间戳降序排序
	sort.Slice(requests, func(i, j int) bool {
		return requests[i].Timestamp.After(requests[j].Timestamp)
	})

	return requests, nil
}

// ClearHistory 清空历史记录
func (hm *HistoryManager) ClearHistory() error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.history = ring.New(hm.size)
	return nil
}
