package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
)

type Storage interface {
	SaveRequest(req *Request) error
	GetRequests(userID string) ([]*Request, error)
	GetAllRequests() ([]*Request, error)
	GetRequestByID(requestID string) (*Request, error)
	DeleteRequest(requestID string) error
}

type FileStorage struct {
	basePath string
}

func NewFileStorage(basePath string) (*FileStorage, error) {
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, err
	}
	return &FileStorage{basePath: basePath}, nil
}

func (s *FileStorage) SaveRequest(req *Request) error {
	userPath := filepath.Join(s.basePath, req.UserID)
	if err := os.MkdirAll(userPath, 0755); err != nil {
		return err
	}

	filename := filepath.Join(userPath, req.ID+".json")
	data, err := json.Marshal(req)
	if err != nil {
		return err
	}

	return os.WriteFile(filename, data, 0644)
}

func (s *FileStorage) GetRequests(userID string) ([]*Request, error) {
	userPath := filepath.Join(s.basePath, userID)
	entries, err := os.ReadDir(userPath)
	if err != nil {
		return nil, err
	}

	var requests []*Request
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		data, err := os.ReadFile(filepath.Join(userPath, entry.Name()))
		if err != nil {
			continue
		}
		var req Request
		if err := json.Unmarshal(data, &req); err != nil {
			continue
		}
		requests = append(requests, &req)
	}
	return requests, nil
}

func (s *FileStorage) GetAllRequests() ([]*Request, error) {
	entries, err := os.ReadDir(s.basePath)
	if err != nil {
		return nil, err
	}

	var allRequests []*Request
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		requests, err := s.GetRequests(entry.Name())
		if err != nil {
			continue
		}
		allRequests = append(allRequests, requests...)
	}
	return allRequests, nil
}

func (s *FileStorage) GetRequestByID(requestID string) (*Request, error) {
	entries, err := os.ReadDir(s.basePath)
	if err != nil {
		return nil, err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		requests, err := s.GetRequests(entry.Name())
		if err != nil {
			continue
		}
		for _, req := range requests {
			if req.ID == requestID {
				return req, nil
			}
		}
	}
	return nil, os.ErrNotExist
}

func (s *FileStorage) DeleteRequest(requestID string) error {
	entries, err := os.ReadDir(s.basePath)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		userPath := filepath.Join(s.basePath, entry.Name())
		userEntries, err := os.ReadDir(userPath)
		if err != nil {
			continue
		}
		for _, userEntry := range userEntries {
			if userEntry.IsDir() {
				continue
			}
			data, err := os.ReadFile(filepath.Join(userPath, userEntry.Name()))
			if err != nil {
				continue
			}
			var req Request
			if err := json.Unmarshal(data, &req); err != nil {
				continue
			}
			if req.ID == requestID {
				return os.Remove(filepath.Join(userPath, userEntry.Name()))
			}
		}
	}
	return os.ErrNotExist
}
