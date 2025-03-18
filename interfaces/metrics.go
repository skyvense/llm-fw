package interfaces

import (
	"llm-fw/api"
)

// MetricsCollector 定义指标收集器的接口
type MetricsCollector interface {
	RecordRequest(model, server string, tokensIn, tokensOut int64, latency int64, isSuccess bool)
	GetModelStats(model string) *api.ModelStats
	GetAllModelStats() map[string]*api.ModelStats
	UpdateServerHealth(server string, isHealthy bool)
	CleanupSystemStats()
}
