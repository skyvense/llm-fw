package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"llm-fw/common"
	"llm-fw/types"
)

// FileStorage 实现了 Storage 接口
type FileStorage struct {
	baseDir string
}

// NewFileStorage 创建一个新的文件存储
func NewFileStorage(baseDir string) (types.Storage, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("创建基础目录失败: %v", err)
	}

	return &FileStorage{
		baseDir: baseDir,
	}, nil
}

// SaveRequest 保存请求记录
func (s *FileStorage) SaveRequest(req *common.Request) error {
	// 确保用户目录存在
	userDir := filepath.Join(s.baseDir, req.UserID)
	if err := os.MkdirAll(userDir, 0755); err != nil {
		return fmt.Errorf("创建用户目录失败: %v", err)
	}

	// 创建请求文件
	filePath := filepath.Join(userDir, fmt.Sprintf("%s.json", req.ID))
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("创建请求文件失败: %v", err)
	}
	defer file.Close()

	// 将请求写入文件
	encoder := json.NewEncoder(file)
	if err := encoder.Encode(req); err != nil {
		return fmt.Errorf("编码请求失败: %v", err)
	}

	return nil
}

// GetModelStats 获取模型统计信息
func (s *FileStorage) GetModelStats(model string) (*types.ModelStats, error) {
	stats := &types.ModelStats{}

	// 遍历所有用户目录
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		return nil, fmt.Errorf("读取基础目录失败: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			userRequests, err := s.GetRequests(entry.Name())
			if err != nil {
				continue
			}

			for _, req := range userRequests {
				if req.Model == model {
					stats.TotalRequests++
					stats.TotalTokensIn += int64(req.TokensIn)
					stats.TotalTokensOut += int64(req.TokensOut)
					stats.AverageLatency += float64(req.LatencyMs)
					if req.Status != 0 {
						stats.FailedRequests++
					}
				}
			}
		}
	}

	// 计算平均延迟
	if stats.TotalRequests > 0 {
		stats.AverageLatency /= float64(stats.TotalRequests)
	}

	return stats, nil
}

// GetAllModelStats 获取所有模型的统计信息
func (s *FileStorage) GetAllModelStats() (map[string]*types.ModelStats, error) {
	stats := make(map[string]*types.ModelStats)

	// 遍历所有用户目录
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		return nil, fmt.Errorf("读取基础目录失败: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			userRequests, err := s.GetRequests(entry.Name())
			if err != nil {
				continue
			}

			for _, req := range userRequests {
				if _, exists := stats[req.Model]; !exists {
					stats[req.Model] = &types.ModelStats{}
				}
				stats[req.Model].TotalRequests++
				stats[req.Model].TotalTokensIn += int64(req.TokensIn)
				stats[req.Model].TotalTokensOut += int64(req.TokensOut)
				stats[req.Model].AverageLatency += float64(req.LatencyMs)
				if req.Status != 0 {
					stats[req.Model].FailedRequests++
				}
			}
		}
	}

	// 计算每个模型的平均延迟
	for _, stat := range stats {
		if stat.TotalRequests > 0 {
			stat.AverageLatency /= float64(stat.TotalRequests)
		}
	}

	return stats, nil
}

// GetRecentRequests 获取最近的请求记录
func (s *FileStorage) GetRecentRequests(limit int) ([]*common.Request, error) {
	var allRequests []*common.Request

	// 遍历所有用户目录
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		return nil, fmt.Errorf("读取基础目录失败: %v", err)
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

	// 限制返回数量
	if len(allRequests) > limit {
		allRequests = allRequests[:limit]
	}

	return allRequests, nil
}

// GetRequests 获取指定用户的所有请求
func (s *FileStorage) GetRequests(userID string) ([]*common.Request, error) {
	userDir := filepath.Join(s.baseDir, userID)
	if _, err := os.Stat(userDir); os.IsNotExist(err) {
		return []*common.Request{}, nil
	}

	files, err := os.ReadDir(userDir)
	if err != nil {
		return nil, fmt.Errorf("读取用户目录失败: %v", err)
	}

	var requests []*common.Request
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			filePath := filepath.Join(userDir, file.Name())
			file, err := os.Open(filePath)
			if err != nil {
				continue
			}

			var req common.Request
			decoder := json.NewDecoder(file)
			if err := decoder.Decode(&req); err != nil {
				file.Close()
				continue
			}
			file.Close()

			requests = append(requests, &req)
		}
	}

	return requests, nil
}

// GetAllRequests 获取所有请求
func (s *FileStorage) GetAllRequests() ([]*common.Request, error) {
	var allRequests []*common.Request

	// 遍历所有用户目录
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		return nil, fmt.Errorf("读取基础目录失败: %v", err)
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

	return allRequests, nil
}

// GetRequestByID 根据ID获取请求
func (s *FileStorage) GetRequestByID(requestID string) (*common.Request, error) {
	// 遍历所有用户目录
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		return nil, fmt.Errorf("读取基础目录失败: %v", err)
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

			var req common.Request
			decoder := json.NewDecoder(file)
			if err := decoder.Decode(&req); err != nil {
				file.Close()
				continue
			}
			file.Close()

			return &req, nil
		}
	}

	return nil, fmt.Errorf("请求未找到")
}

// DeleteRequest 删除请求
func (s *FileStorage) DeleteRequest(requestID string) error {
	// 遍历所有用户目录
	entries, err := os.ReadDir(s.baseDir)
	if err != nil {
		return fmt.Errorf("读取基础目录失败: %v", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			filePath := filepath.Join(s.baseDir, entry.Name(), fmt.Sprintf("%s.json", requestID))
			if _, err := os.Stat(filePath); os.IsNotExist(err) {
				continue
			}

			if err := os.Remove(filePath); err != nil {
				return fmt.Errorf("删除请求文件失败: %v", err)
			}

			return nil
		}
	}

	return fmt.Errorf("请求未找到")
}

// Close 关闭存储连接
func (s *FileStorage) Close() error {
	return nil
}

// NewHistoryManager 创建一个新的历史记录管理器
func (s *FileStorage) NewHistoryManager(size int) types.HistoryManager {
	return &FileHistoryManager{
		storage: s,
		size:    size,
		entries: make([]*common.Request, 0, size),
	}
}

// FileHistoryManager 实现了 HistoryManager 接口
type FileHistoryManager struct {
	storage types.Storage
	size    int
	entries []*common.Request
}

// Add 添加一条历史记录
func (h *FileHistoryManager) Add(req *common.Request) {
	h.entries = append(h.entries, req)
	if len(h.entries) > h.size {
		h.entries = h.entries[1:]
	}
}

// Get 获取所有历史记录
func (h *FileHistoryManager) Get() []*common.Request {
	return h.entries
}

// Clear 清空历史记录
func (h *FileHistoryManager) Clear() {
	h.entries = make([]*common.Request, 0, h.size)
}
