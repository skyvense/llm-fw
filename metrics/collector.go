package metrics

import (
	"sync"
	"time"

	"llm-fw/types"
)

// Metrics 管理所有指标
type Metrics struct {
	mu             sync.RWMutex
	ModelStats     map[string]*types.ModelStats
	serverHealth   map[string]bool
	totalRequests  int64
	totalTokensIn  int64
	totalTokensOut int64
	failedRequests int64
}

// NewMetrics 创建一个新的指标收集器
func NewMetrics() *Metrics {
	return &Metrics{
		ModelStats:   make(map[string]*types.ModelStats),
		serverHealth: make(map[string]bool),
	}
}

// RecordRequest 记录一个请求的统计信息
func (m *Metrics) RecordRequest(model, server string, tokensIn, tokensOut int64, latency int64, isSuccess bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	stats, exists := m.ModelStats[model]
	if !exists {
		stats = &types.ModelStats{
			TotalRequests:  0,
			TotalTokensIn:  0,
			TotalTokensOut: 0,
			AverageLatency: 0,
			FailedRequests: 0,
			LastUsed:       time.Now(),
		}
		m.ModelStats[model] = stats
	}

	stats.TotalRequests++
	stats.TotalTokensIn += tokensIn
	stats.TotalTokensOut += tokensOut
	stats.LastUsed = time.Now()

	// 更新平均延迟
	if stats.TotalRequests == 1 {
		stats.AverageLatency = float64(latency)
	} else {
		stats.AverageLatency = (stats.AverageLatency*float64(stats.TotalRequests-1) + float64(latency)) / float64(stats.TotalRequests)
	}

	if !isSuccess {
		stats.FailedRequests++
	}

	// 更新总体统计信息
	m.totalRequests++
	m.totalTokensIn += tokensIn
	m.totalTokensOut += tokensOut
	if !isSuccess {
		m.failedRequests++
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

	stats := make(map[string]*types.ModelStats)
	for model, modelStats := range m.ModelStats {
		stats[model] = modelStats
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
