package storage

import (
	"llm-fw/handlers"
)

// StorageAdapter adapts FileStorage to handlers.Storage interface
type StorageAdapter struct {
	storage *FileStorage
}

// NewStorageAdapter creates a new storage adapter
func NewStorageAdapter(storage *FileStorage) *StorageAdapter {
	return &StorageAdapter{storage: storage}
}

func (a *StorageAdapter) SaveRequest(req *handlers.Request) error {
	return a.storage.SaveRequest(FromHandlerRequest(req))
}

func (a *StorageAdapter) GetRequests(userID string) ([]*handlers.Request, error) {
	requests, err := a.storage.GetRequests(userID)
	if err != nil {
		return nil, err
	}

	handlerRequests := make([]*handlers.Request, len(requests))
	for i, req := range requests {
		handlerRequests[i] = req.ToHandlerRequest()
	}
	return handlerRequests, nil
}

func (a *StorageAdapter) GetAllRequests() ([]*handlers.Request, error) {
	requests, err := a.storage.GetAllRequests()
	if err != nil {
		return nil, err
	}

	handlerRequests := make([]*handlers.Request, len(requests))
	for i, req := range requests {
		handlerRequests[i] = req.ToHandlerRequest()
	}
	return handlerRequests, nil
}

func (a *StorageAdapter) GetRequestByID(requestID string) (*handlers.Request, error) {
	request, err := a.storage.GetRequestByID(requestID)
	if err != nil {
		return nil, err
	}
	return request.ToHandlerRequest(), nil
}

func (a *StorageAdapter) DeleteRequest(requestID string) error {
	return a.storage.DeleteRequest(requestID)
}
