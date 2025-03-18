package handlers

import "llm-fw/types"

// MetricsCollector 定义了指标收集器的接口
type MetricsCollector interface {
	RecordRequest(model, server string, tokensIn, tokensOut int64, latency int64, isSuccess bool)
	GetMetrics() *types.Metrics
	UpdateServerHealth(server string, isHealthy bool)
}
