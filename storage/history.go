package storage

import (
	"container/ring"
	"sort"
	"strings"
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

// Search 搜索历史记录
func (hm *HistoryManager) Search(params types.SearchParams) (*types.SearchResult, error) {
	// 从存储中获取所有请求
	requests, err := hm.storage.GetAllRequests()
	if err != nil {
		return nil, err
	}

	// 应用过滤条件
	var filtered []*common.Request
	for _, req := range requests {
		// 检查模型
		if params.Model != "" && req.Model != params.Model {
			continue
		}

		// 检查日期范围
		if !params.StartDate.IsZero() && req.Timestamp.Before(params.StartDate) {
			continue
		}
		if !params.EndDate.IsZero() && req.Timestamp.After(params.EndDate) {
			continue
		}

		// 检查关键词
		if params.Keyword != "" {
			keyword := strings.ToLower(params.Keyword)
			prompt := strings.ToLower(req.Prompt)
			response := strings.ToLower(req.Response)
			if !strings.Contains(prompt, keyword) && !strings.Contains(response, keyword) {
				continue
			}
		}

		filtered = append(filtered, req)
	}

	// 按时间戳降序排序
	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].Timestamp.After(filtered[j].Timestamp)
	})

	// 计算总记录数
	total := len(filtered)

	// 应用分页
	start := (params.Page - 1) * params.PageSize
	end := start + params.PageSize
	if start >= len(filtered) {
		filtered = []*common.Request{}
	} else {
		if end > len(filtered) {
			end = len(filtered)
		}
		filtered = filtered[start:end]
	}

	return &types.SearchResult{
		Requests: filtered,
		Total:    total,
	}, nil
}

// Clear 清空历史记录
func (hm *HistoryManager) Clear() {
	hm.mu.Lock()
	defer hm.mu.Unlock()

	hm.history = ring.New(hm.size)
}
