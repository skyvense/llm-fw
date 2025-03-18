package types

import "llm-fw/common"

// ModelStats 存储模型的统计信息
type ModelStats = common.ModelStats

// ModelInfo 表示模型的完整信息
type ModelInfo = common.ModelInfo

// RequestStats 包含请求的统计信息
type RequestStats = common.RequestStats

// HistoryEntry 表示一条历史记录
type HistoryEntry = common.HistoryEntry

// Request 表示一个请求
type Request = common.Request

// Metrics 表示指标数据
type Metrics struct {
	TotalRequests  int64
	TotalTokensIn  int64
	TotalTokensOut int64
	TotalLatencyMs int64
	FailedRequests int64
	ServerHealth   map[string]bool
	ModelStats     map[string]*ModelStats
}

// MetricsCollector 定义了指标收集器的接口
type MetricsCollector interface {
	RecordRequest(model, server string, tokensIn, tokensOut int64, latency int64, isSuccess bool)
	GetMetrics() *Metrics
	UpdateServerHealth(server string, isHealthy bool)
	CleanupSystemStats()
	GetAllModelStats() map[string]*ModelStats
}
