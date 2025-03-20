package common

import "time"

// ModelStats 存储模型的统计信息
type ModelStats struct {
	TotalRequests   int64     `json:"total_requests"`
	TotalTokensIn   int64     `json:"total_tokens_in"`
	TotalTokensOut  int64     `json:"total_tokens_out"`
	AverageLatency  float64   `json:"average_latency"`
	TokensPerSecond float64   `json:"tokens_per_second"` // 每秒生成的token数量
	FailedRequests  int64     `json:"failed_requests"`
	LastUsed        time.Time `json:"last_used"`
}

// ModelInfo 表示模型的完整信息
type ModelInfo struct {
	Name        string      `json:"name"`
	Family      string      `json:"family,omitempty"`
	Parameters  string      `json:"parameters,omitempty"`
	Format      string      `json:"format,omitempty"`
	Stats       *ModelStats `json:"stats,omitempty"`
	LastUsed    *time.Time  `json:"last_used,omitempty"`
	IsAvailable bool        `json:"is_available"`
}

// RequestStats 包含请求的统计信息
type RequestStats struct {
	TokensIn  int     `json:"tokens_in"`
	TokensOut int     `json:"tokens_out"`
	LatencyMs float64 `json:"latency_ms"`
}

// HistoryEntry 表示一条历史记录
type HistoryEntry struct {
	Model     string       `json:"model"`
	Prompt    string       `json:"prompt"`
	Response  string       `json:"response"`
	Stats     RequestStats `json:"stats"`
	Timestamp time.Time    `json:"timestamp"`
}

// Request 表示一个请求
type Request struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Model     string    `json:"model"`
	Prompt    string    `json:"prompt"`
	Response  string    `json:"response"`
	TokensIn  int       `json:"tokens_in"`
	TokensOut int       `json:"tokens_out"`
	Server    string    `json:"server"`
	LatencyMs float64   `json:"latency_ms"`
	Status    int       `json:"status"` // 0: 成功, 1: 失败
	Error     string    `json:"error"`  // 错误信息
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"` // 请求来源：internal_ui, external_ui, api
}

// ModelStatsHistory 表示模型统计历史记录
type ModelStatsHistory struct {
	ID              string    `json:"id"`
	Model           string    `json:"model"`
	TotalRequests   int64     `json:"total_requests"`
	FailedRequests  int64     `json:"failed_requests"`
	TotalTokensIn   int64     `json:"total_tokens_in"`
	TotalTokensOut  int64     `json:"total_tokens_out"`
	AverageLatency  float64   `json:"average_latency"`
	TokensPerSecond float64   `json:"tokens_per_second"` // 每秒生成的token数量
	Timestamp       time.Time `json:"timestamp"`
}
