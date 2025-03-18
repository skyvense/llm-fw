package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"llm-fw/api"
	"llm-fw/interfaces"
)

// Storage 定义了存储接口
type Storage interface {
	SaveRequest(req *api.Request) error
	GetRequests(userID string) ([]*api.Request, error)
	GetAllRequests() ([]*api.Request, error)
	GetRequestByID(requestID string) (*api.Request, error)
	DeleteRequest(requestID string) error
	NewHistoryManager(size int) interfaces.HistoryManager
}

// FileStorage 实现了 Storage 接口
type FileStorage struct {
	baseDir string
}

// NewFileStorage 创建一个新的文件存储
func NewFileStorage(baseDir string) (*FileStorage, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %v", err)
	}

	return &FileStorage{
		baseDir: baseDir,
	}, nil
}

// NewHistoryManager 创建一个新的历史记录管理器
func (s *FileStorage) NewHistoryManager(size int) interfaces.HistoryManager {
	return NewHistoryManager(s, size)
}

// SaveRequest 保存请求到文件
func (s *FileStorage) SaveRequest(req *api.Request) error {
	// 确保用户目录存在
	userDir := filepath.Join(s.baseDir, req.UserID)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return fmt.Errorf("failed to create user directory: %v", err)
	}

	// 创建请求文件
	filePath := filepath.Join(userDir, fmt.Sprintf("%s.json", req.ID))
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create request file: %v", err)
	}
	defer file.Close()

	// 将请求写入文件
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(req); err != nil {
		return fmt.Errorf("failed to encode request: %v", err)
	}

	return nil
}

// GetRequests 获取指定用户的所有请求
func (s *FileStorage) GetRequests(userID string) ([]*api.Request, error) {
	userDir := filepath.Join(s.baseDir, userID)
	if _, err := os.Stat(userDir); os.IsNotExist(err) {
		return []*api.Request{}, nil
	}

	files, err := os.ReadDir(userDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read user directory: %v", err)
	}

	var requests []*api.Request
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			filePath := filepath.Join(userDir, file.Name())
			file, err := os.Open(filePath)
			if err != nil {
				continue
			}

			var req api.Request
			decoder := json.NewDecoder(file)
			if err := decoder.Decode(&req); err != nil {
				file.Close()
				continue
			}
			file.Close()

			requests = append(requests, &req)
		}
	}

	// 按时间戳排序
	sort.Slice(requests, func(i, j int) bool {
		return requests[i].Timestamp.After(requests[j].Timestamp)
	})

	return requests, nil
}

// GetAllRequests 获取所有请求
func (s *FileStorage) GetAllRequests() ([]*api.Request, error) {
	var allRequests []*api.Request

	// 遍历所有用户目录
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read base directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			userRequests, err := s.GetRequests(entry.Name())
			if err != nil {
				continue
			}
			allRequests = append(allRequests, userRequests...)
		}
	}

	// 按时间戳排序
	sort.Slice(allRequests, func(i, j int) bool {
		return allRequests[i].Timestamp.After(allRequests[j].Timestamp)
	})

	return allRequests, nil
}

// GetRequestByID 根据ID获取请求
func (s *FileStorage) GetRequestByID(requestID string) (*api.Request, error) {
	// 遍历所有用户目录
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read base directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			filePath := filepath.Join(s.baseDir, entry.Name(), fmt.Sprintf("%s.json", requestID))
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				continue
			}

			file, err := os.Open(filePath)
			if err != nil {
				continue
			}

			var req api.Request
			decoder := json.NewDecoder(file)
			if err := decoder.Decode(&req); err != nil {
				file.Close()
				continue
			}
			file.Close()

			return &req, nil
		}
	}

	return nil, fmt.Errorf("request not found")
}

// DeleteRequest 删除请求
func (s *FileStorage) DeleteRequest(requestID string) error {
	// 遍历所有用户目录
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		return fmt.Errorf("failed to read base directory: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			filePath := filepath.Join(s.baseDir, entry.Name(), fmt.Sprintf("%s.json", requestID))
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				continue
			}

			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("failed to delete request file: %v", err)
			}

			return nil
		}
	}

	return fmt.Errorf("request not found")
}
