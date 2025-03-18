package storage

import (
	"container/ring"
	"sort"
	"sync"
	"time"

	"llm-fw/common"
	"llm-fw/types"
)

// HistoryManager 管理请求历史记录
type HistoryManager struct {
	storage types.Storage
	history *ring.Ring
	mu      sync.RWMutex
	size    int
}

// NewHistoryManager 创建一个新的历史记录管理器
func NewHistoryManager(storage types.Storage, size int) *HistoryManager {
	return &HistoryManager{
		storage: storage,
		history: ring.New(size),
		size:    size,
	}
}

// Add 添加一条新的历史记录
func (hm *HistoryManager) Add(req *common.Request) {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	req.Timestamp = time.Now()
	hm.history.Value = req
	hm.history = hm.history.Next()
}

// Get 获取历史记录
func (hm *HistoryManager) Get() []*common.Request {
	// 从存储中获取所有请求
	requests, err := hm.storage.GetAllRequests()
	if err != nil {
		return nil
	}

	// 按时间戳降序排序
	sort.Slice(requests, func(i, j int) bool {
		return requests[i].Timestamp.After(requests[j].Timestamp)
	})

	return requests
}

// Clear 清空历史记录
func (hm *HistoryManager) Clear() {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.history = ring.New(hm.size)
}
