package types

// NoopMetricsCollector 是一个空的指标收集器实现
type NoopMetricsCollector struct{}

// RecordRequest 实现了 MetricsCollector 接口
func (c *NoopMetricsCollector) RecordRequest(model string, server string, tokensIn int64, tokensOut int64, latencyMs int64, success bool) {
	// 什么都不做
}

// GetMetrics 实现了 MetricsCollector 接口
func (c *NoopMetricsCollector) GetMetrics() *Metrics {
	return &Metrics{
		ServerHealth: make(map[string]bool),
		ModelStats:   make(map[string]*ModelStats),
	}
}

// UpdateServerHealth 实现了 MetricsCollector 接口
func (c *NoopMetricsCollector) UpdateServerHealth(server string, isHealthy bool) {
	// 什么都不做
}

// CleanupSystemStats 实现了 MetricsCollector 接口
func (c *NoopMetricsCollector) CleanupSystemStats() {
	// 什么都不做
}

// GetAllModelStats 实现了 MetricsCollector 接口
func (c *NoopMetricsCollector) GetAllModelStats() map[string]*ModelStats {
	return make(map[string]*ModelStats)
}
