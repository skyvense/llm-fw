package interfaces

import (
	"llm-fw/api"
)

// Storage 定义了存储接口
type Storage interface {
	SaveRequest(req *api.Request) error
	GetRequests(userID string) ([]*api.Request, error)
	GetAllRequests() ([]*api.Request, error)
	GetRequestByID(requestID string) (*api.Request, error)
	DeleteRequest(requestID string) error
	NewHistoryManager(size int) HistoryManager
}
