package handlers

import "llm-fw/metrics"

// MetricsCollector 定义了指标收集器的接口
type MetricsCollector interface {
	RecordRequest(model, server string, tokensIn, tokensOut int64, latency int64, isSuccess bool)
	GetMetrics() *metrics.Metrics
	UpdateServerHealth(server string, isHealthy bool)
}

// Storage 定义了存储接口
type Storage interface {
	SaveRequest(req *Request) error
	GetRequests(userID string) ([]*Request, error)
	GetAllRequests() ([]*Request, error)
	GetRequestByID(requestID string) (*Request, error)
	DeleteRequest(requestID string) error
}

// Request 定义了请求记录的结构
type Request struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Model     string `json:"model"`
	Prompt    string `json:"prompt"`
	Response  string `json:"response"`
	TokensIn  int    `json:"tokens_in"`
	TokensOut int    `json:"tokens_out"`
	Server    string `json:"server"`
}
