package metrics

import (
	"sync"
	"time"
)

type Metrics struct {
	mu sync.RWMutex

	// 请求统计
	TotalRequests  int64
	TotalTokensIn  int64
	TotalTokensOut int64
	FailedRequests int64
	AverageLatency float64
	TotalLatency   int64

	// 模型统计
	ModelStats map[string]*ModelMetrics

	// 服务器统计
	ServerStats map[string]*ServerMetrics

	// 实时统计（最近1小时）
	RecentRequests []*RequestMetrics
}

type ModelMetrics struct {
	TotalRequests  int64
	TotalTokensIn  int64
	TotalTokensOut int64
	FailedRequests int64
	AverageLatency float64
	TotalLatency   int64
}

type ServerMetrics struct {
	TotalRequests  int64
	TotalTokensIn  int64
	TotalTokensOut int64
	FailedRequests int64
	AverageLatency float64
	TotalLatency   int64
	IsHealthy      bool
	LastCheck      time.Time
}

type RequestMetrics struct {
	Timestamp time.Time
	Model     string
	Server    string
	TokensIn  int64
	TokensOut int64
	Latency   int64
	IsSuccess bool
}

func NewMetrics() *Metrics {
	return &Metrics{
		ModelStats:     make(map[string]*ModelMetrics),
		ServerStats:    make(map[string]*ServerMetrics),
		RecentRequests: make([]*RequestMetrics, 0),
	}
}

func (m *Metrics) RecordRequest(model, server string, tokensIn, tokensOut int64, latency int64, isSuccess bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 更新总体统计
	m.TotalRequests++
	m.TotalTokensIn += tokensIn
	m.TotalTokensOut += tokensOut
	m.TotalLatency += latency
	m.AverageLatency = float64(m.TotalLatency) / float64(m.TotalRequests)
	if !isSuccess {
		m.FailedRequests++
	}

	// 更新模型统计
	if _, exists := m.ModelStats[model]; !exists {
		m.ModelStats[model] = &ModelMetrics{}
	}
	modelStats := m.ModelStats[model]
	modelStats.TotalRequests++
	modelStats.TotalTokensIn += tokensIn
	modelStats.TotalTokensOut += tokensOut
	modelStats.TotalLatency += latency
	modelStats.AverageLatency = float64(modelStats.TotalLatency) / float64(modelStats.TotalRequests)
	if !isSuccess {
		modelStats.FailedRequests++
	}

	// 更新服务器统计
	if _, exists := m.ServerStats[server]; !exists {
		m.ServerStats[server] = &ServerMetrics{}
	}
	serverStats := m.ServerStats[server]
	serverStats.TotalRequests++
	serverStats.TotalTokensIn += tokensIn
	serverStats.TotalTokensOut += tokensOut
	serverStats.TotalLatency += latency
	serverStats.AverageLatency = float64(serverStats.TotalLatency) / float64(serverStats.TotalRequests)
	if !isSuccess {
		serverStats.FailedRequests++
	}

	// 更新实时统计
	now := time.Now()
	recentMetrics := &RequestMetrics{
		Timestamp: now,
		Model:     model,
		Server:    server,
		TokensIn:  tokensIn,
		TokensOut: tokensOut,
		Latency:   latency,
		IsSuccess: isSuccess,
	}
	m.RecentRequests = append(m.RecentRequests, recentMetrics)

	// 清理超过1小时的数据
	cutoff := now.Add(-1 * time.Hour)
	valid := m.RecentRequests[:0]
	for _, req := range m.RecentRequests {
		if req.Timestamp.After(cutoff) {
			valid = append(valid, req)
		}
	}
	m.RecentRequests = valid
}

func (m *Metrics) UpdateServerHealth(server string, isHealthy bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if stats, exists := m.ServerStats[server]; exists {
		stats.IsHealthy = isHealthy
		stats.LastCheck = time.Now()
	}
}

func (m *Metrics) GetMetrics() *Metrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 创建副本以避免并发访问问题
	metrics := &Metrics{
		TotalRequests:  m.TotalRequests,
		TotalTokensIn:  m.TotalTokensIn,
		TotalTokensOut: m.TotalTokensOut,
		FailedRequests: m.FailedRequests,
		AverageLatency: m.AverageLatency,
		TotalLatency:   m.TotalLatency,
		ModelStats:     make(map[string]*ModelMetrics),
		ServerStats:    make(map[string]*ServerMetrics),
		RecentRequests: make([]*RequestMetrics, len(m.RecentRequests)),
	}

	// 复制模型统计
	for k, v := range m.ModelStats {
		metrics.ModelStats[k] = &ModelMetrics{
			TotalRequests:  v.TotalRequests,
			TotalTokensIn:  v.TotalTokensIn,
			TotalTokensOut: v.TotalTokensOut,
			FailedRequests: v.FailedRequests,
			AverageLatency: v.AverageLatency,
			TotalLatency:   v.TotalLatency,
		}
	}

	// 复制服务器统计
	for k, v := range m.ServerStats {
		metrics.ServerStats[k] = &ServerMetrics{
			TotalRequests:  v.TotalRequests,
			TotalTokensIn:  v.TotalTokensIn,
			TotalTokensOut: v.TotalTokensOut,
			FailedRequests: v.FailedRequests,
			AverageLatency: v.AverageLatency,
			TotalLatency:   v.TotalLatency,
			IsHealthy:      v.IsHealthy,
			LastCheck:      v.LastCheck,
		}
	}

	// 复制实时统计
	copy(metrics.RecentRequests, m.RecentRequests)

	return metrics
}
