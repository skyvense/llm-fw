package types

// Storage 定义了存储接口
type Storage interface {
	// SaveRequest 保存请求记录
	SaveRequest(req *Request) error

	// GetModelStats 获取模型统计信息
	GetModelStats(model string) (*ModelStats, error)

	// GetAllModelStats 获取所有模型的统计信息
	GetAllModelStats() (map[string]*ModelStats, error)

	// GetRecentRequests 获取最近的请求记录
	GetRecentRequests(limit int) ([]*Request, error)

	// GetRequests 获取指定用户的所有请求
	GetRequests(userID string) ([]*Request, error)

	// GetAllRequests 获取所有请求
	GetAllRequests() ([]*Request, error)

	// GetRequestByID 根据ID获取请求
	GetRequestByID(requestID string) (*Request, error)

	// DeleteRequest 删除请求
	DeleteRequest(requestID string) error

	// NewHistoryManager 创建一个新的历史记录管理器
	NewHistoryManager(size int) HistoryManager

	// Close 关闭存储连接
	Close() error

	// GetModelStatsHistory 获取模型统计历史
	GetModelStatsHistory(model string, limit int) ([]*ModelStatsHistory, error)

	// ListModelStatsHistory 获取所有模型统计历史
	ListModelStatsHistory(limit int) ([]*ModelStatsHistory, error)

	// SaveModelStats 保存模型统计信息
	SaveModelStats(model string, stats *ModelStats) error

	// SaveModelStatsHistory 保存模型统计历史
	SaveModelStatsHistory(history *ModelStatsHistory) error

	// ListRequests 获取请求列表
	ListRequests(limit int) ([]*Request, error)
}

// HistoryManager 定义了历史记录管理器的接口
type HistoryManager interface {
	Add(req *Request)
	Get() []*Request
	Clear()
}
