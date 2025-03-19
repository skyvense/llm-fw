package interfaces

import "llm-fw/types"

// HistoryManager 定义了历史记录管理器的接口
type HistoryManager interface {
	// AddEntry 添加一条新的历史记录
	AddEntry(entry types.Request) error
	// GetHistory 获取历史记录
	GetHistory() ([]*types.Request, error)
	// ClearHistory 清空历史记录
	ClearHistory() error
}
