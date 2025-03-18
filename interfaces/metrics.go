package interfaces

import (
	"llm-fw/common"
	"llm-fw/metrics"
)

// MetricsCollector 定义指标收集器的接口
type MetricsCollector interface {
	RecordRequest(model, server string, tokensIn, tokensOut int64, latency int64, isSuccess bool)
	GetModelStats(model string) *common.ModelStats
	GetAllModelStats() map[string]*common.ModelStats
	UpdateServerHealth(server string, isHealthy bool)
	CleanupSystemStats()
	GetMetrics() *metrics.Metrics
}
