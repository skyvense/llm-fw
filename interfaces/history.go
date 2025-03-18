package interfaces

import "llm-fw/api"

// HistoryManager 定义历史记录管理器的接口
type HistoryManager interface {
	AddEntry(entry api.HistoryEntry) error
	GetHistory() ([]api.HistoryEntry, error)
	ClearHistory() error
}
