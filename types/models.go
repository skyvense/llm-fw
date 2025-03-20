package types

import (
	"llm-fw/common"
	"time"
)

// ModelStats 存储模型的统计信息
type ModelStats = common.ModelStats

// ModelInfo 表示模型的完整信息
type ModelInfo struct {
	Name        string             `json:"name"`
	Family      string             `json:"family,omitempty"`
	Parameters  string             `json:"parameters,omitempty"`
	Format      string             `json:"format,omitempty"`
	Stats       *ModelStats        `json:"stats,omitempty"`
	History     *ModelStatsHistory `json:"history,omitempty"`
	LastUsed    *time.Time         `json:"last_used,omitempty"`
	IsAvailable bool               `json:"is_available"`
}

// RequestStats 包含请求的统计信息
type RequestStats = common.RequestStats

// HistoryEntry 表示一条历史记录
type HistoryEntry = common.HistoryEntry

// Request 表示一个请求
type Request = common.Request

// ModelStatsHistory 表示模型统计历史记录
type ModelStatsHistory = common.ModelStatsHistory

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
