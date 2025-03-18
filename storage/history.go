package storage

import (
	"container/ring"
	"sync"
	"time"

	"llm-fw/api"
)

// HistoryManager 管理请求历史记录
type HistoryManager struct {
	history *ring.Ring
	mu      sync.RWMutex
	size    int
}

// NewHistoryManager 创建一个新的历史记录管理器
func NewHistoryManager(size int) *HistoryManager {
	return &HistoryManager{
		history: ring.New(size),
		size:    size,
	}
}

// AddEntry 添加一条新的历史记录
func (hm *HistoryManager) AddEntry(entry api.HistoryEntry) error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	entry.Timestamp = time.Now()
	hm.history.Value = entry
	hm.history = hm.history.Next()

	return nil
}

// GetHistory 获取历史记录
func (hm *HistoryManager) GetHistory() ([]api.HistoryEntry, error) {
	hm.mu.RLock()
	defer hm.mu.RUnlock()

	entries := make([]api.HistoryEntry, 0, hm.size)
	current := hm.history.Prev()

	for i := 0; i < hm.size; i++ {
		if current.Value == nil {
			break
		}
		entries = append(entries, current.Value.(api.HistoryEntry))
		current = current.Prev()
	}

	return entries, nil
}

// ClearHistory 清空历史记录
func (hm *HistoryManager) ClearHistory() error {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.history = ring.New(hm.size)
	return nil
}
