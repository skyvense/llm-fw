package metrics

import (
	"log"
	"sync"
	"time"

	"llm-fw/types"
)

// ModelStatsStorage 定义了存储模型统计信息的接口
type ModelStatsStorage interface {
	SaveModelStats(model string, stats *types.ModelStats) error
	GetModelStats(model string) (*types.ModelStats, error)
	GetAllModelStats() (map[string]*types.ModelStats, error)
	DeleteModelStats(model string) error
}

// Metrics 管理所有指标
type Metrics struct {
	mu             sync.RWMutex
	ModelStats     map[string]*types.ModelStats
	serverHealth   map[string]bool
	totalRequests  int64
	totalTokensIn  int64
	totalTokensOut int64
	failedRequests int64
	storage        ModelStatsStorage
}

// NewMetrics 创建一个新的指标收集器
func NewMetrics(storage ModelStatsStorage) *Metrics {
	m := &Metrics{
		ModelStats:   make(map[string]*types.ModelStats),
		serverHealth: make(map[string]bool),
		storage:      storage,
	}

	// 从存储中加载统计信息
	if storage != nil {
		if stats, err := storage.GetAllModelStats(); err == nil {
			m.ModelStats = stats
		} else {
			log.Printf("Failed to load model stats from storage: %v", err)
		}
	}

	return m
}

// RecordRequest 记录一个请求的统计信息
func (m *Metrics) RecordRequest(req *types.Request) {
	m.mu.Lock()
	defer m.mu.Unlock()

	stats, exists := m.ModelStats[req.Model]
	if !exists {
		stats = &types.ModelStats{
			TotalRequests:  0,
			TotalTokensIn:  0,
			TotalTokensOut: 0,
			AverageLatency: 0,
			FailedRequests: 0,
			LastUsed:       time.Now(),
		}
		m.ModelStats[req.Model] = stats
	}

	stats.TotalRequests++
	stats.TotalTokensIn += int64(req.TokensIn)
	stats.TotalTokensOut += int64(req.TokensOut)
	stats.LastUsed = time.Now()

	// 更新平均延迟
	if stats.TotalRequests == 1 {
		stats.AverageLatency = req.LatencyMs
	} else {
		stats.AverageLatency = (stats.AverageLatency*float64(stats.TotalRequests-1) + req.LatencyMs) / float64(stats.TotalRequests)
	}

	if req.Status != 0 { // 0 表示成功
		stats.FailedRequests++
	}

	// 更新总体统计信息
	m.totalRequests++
	m.totalTokensIn += int64(req.TokensIn)
	m.totalTokensOut += int64(req.TokensOut)
	if req.Status != 0 {
		m.failedRequests++
	}

	// 保存统计信息到存储
	if m.storage != nil {
		if err := m.storage.SaveModelStats(req.Model, stats); err != nil {
			log.Printf("Failed to save model stats: %v", err)
		}
	}
}

// GetMetrics 获取所有指标
func (m *Metrics) GetMetrics() *types.Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return &types.Metrics{
		TotalRequests:  m.totalRequests,
		TotalTokensIn:  m.totalTokensIn,
		TotalTokensOut: m.totalTokensOut,
		FailedRequests: m.failedRequests,
		ServerHealth:   m.serverHealth,
		ModelStats:     m.ModelStats,
	}
}

// UpdateServerHealth 更新服务器健康状态
func (m *Metrics) UpdateServerHealth(server string, isHealthy bool) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.serverHealth[server] = isHealthy
}

// GetModelStats 获取指定模型的统计信息
func (m *Metrics) GetModelStats(model string) *types.ModelStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if stats, exists := m.ModelStats[model]; exists {
		return stats
	}
	return nil
}

// GetAllModelStats 获取所有模型的统计信息
func (m *Metrics) GetAllModelStats() map[string]*types.ModelStats {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 从内存中获取最新的统计信息
	stats := make(map[string]*types.ModelStats)
	for model, modelStats := range m.ModelStats {
		// 创建一个新的副本以避免并发问题
		statsCopy := *modelStats
		stats[model] = &statsCopy
	}
	return stats
}

// CleanupSystemStats 清理系统统计信息
func (m *Metrics) CleanupSystemStats() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.totalRequests = 0
	m.totalTokensIn = 0
	m.totalTokensOut = 0
	m.failedRequests = 0
	m.serverHealth = make(map[string]bool)
	m.ModelStats = make(map[string]*types.ModelStats)
}

// DeleteModelStats 删除指定模型的统计数据
func (m *Metrics) DeleteModelStats(modelName string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 从内存中删除统计数据
	delete(m.ModelStats, modelName)

	// 如果有存储接口，也从存储中删除
	if m.storage != nil {
		if err := m.storage.DeleteModelStats(modelName); err != nil {
			log.Printf("Failed to delete model stats from storage: %v", err)
		}
	}
}
