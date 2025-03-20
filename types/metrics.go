package types

// MetricsCollector defines the interface for collecting model metrics
type MetricsCollector interface {
	// 记录请求统计信息
	RecordRequest(req *Request)
	// 获取所有模型的统计信息
	GetAllModelStats() map[string]*ModelStats
	// 删除指定模型的统计信息
	DeleteModelStats(modelName string)
}

// NoopMetricsCollector implements MetricsCollector with no-op operations
type NoopMetricsCollector struct{}

// RecordRequest 实现了 MetricsCollector 接口
func (c *NoopMetricsCollector) RecordRequest(req *Request) {
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

// DeleteModelStats 实现了 MetricsCollector 接口
func (c *NoopMetricsCollector) DeleteModelStats(modelName string) {
	// 什么都不做
}
