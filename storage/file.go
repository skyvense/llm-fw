package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"llm-fw/types"
)

// FileStorageImpl implements Storage interface using files
type FileStorageImpl struct {
	baseDir      string
	mu           sync.RWMutex
	modelStats   map[string]*types.ModelStats
	modelHistory map[string][]*types.ModelStatsHistory
}

// NewFileStorageImpl creates a new FileStorage instance
func NewFileStorageImpl(baseDir string) (*FileStorageImpl, error) {
	if err := os.MkdirAll(baseDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %v", err)
	}

	fs := &FileStorageImpl{
		baseDir:      baseDir,
		modelStats:   make(map[string]*types.ModelStats),
		modelHistory: make(map[string][]*types.ModelStatsHistory),
	}

	if err := fs.loadModelStats(); err != nil {
		return nil, fmt.Errorf("failed to load model stats: %w", err)
	}

	if err := fs.loadModelHistory(); err != nil {
		return nil, fmt.Errorf("failed to load model history: %w", err)
	}

	return fs, nil
}

// loadModelStats loads model statistics from file
func (fs *FileStorageImpl) loadModelStats() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	data, err := os.ReadFile(filepath.Join(fs.baseDir, "model_stats.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return json.Unmarshal(data, &fs.modelStats)
}

// saveModelStats saves model statistics to file
func (fs *FileStorageImpl) saveModelStats() error {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	data, err := json.MarshalIndent(fs.modelStats, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(fs.baseDir, "model_stats.json"), data, 0644)
}

// loadModelHistory loads model history from file
func (fs *FileStorageImpl) loadModelHistory() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	data, err := os.ReadFile(filepath.Join(fs.baseDir, "model_history.json"))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	return json.Unmarshal(data, &fs.modelHistory)
}

// saveModelHistory saves model history to file
func (fs *FileStorageImpl) saveModelHistory() error {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	data, err := json.MarshalIndent(fs.modelHistory, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filepath.Join(fs.baseDir, "model_history.json"), data, 0644)
}

// SaveRequest saves a request
func (fs *FileStorageImpl) SaveRequest(req *types.Request) error {
	// Not implemented
	return nil
}

// GetRequest retrieves a request by ID
func (fs *FileStorageImpl) GetRequest(id string) (*types.Request, error) {
	// Not implemented
	return nil, nil
}

// ListRequests retrieves requests with limit
func (fs *FileStorageImpl) ListRequests(limit int) ([]*types.Request, error) {
	// Not implemented
	return nil, nil
}

// DeleteRequest deletes a request by ID
func (fs *FileStorageImpl) DeleteRequest(id string) error {
	// Not implemented
	return nil
}

// SaveModelStats saves model statistics
func (fs *FileStorageImpl) SaveModelStats(model string, stats *types.ModelStats) error {
	fs.mu.Lock()
	fs.modelStats[model] = stats
	fs.mu.Unlock()

	return fs.saveModelStats()
}

// GetModelStats retrieves model statistics
func (fs *FileStorageImpl) GetModelStats(model string) (*types.ModelStats, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	stats, exists := fs.modelStats[model]
	if !exists {
		return nil, fmt.Errorf("model stats not found: %s", model)
	}

	return stats, nil
}

// GetAllModelStats retrieves all model statistics
func (fs *FileStorageImpl) GetAllModelStats() (map[string]*types.ModelStats, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	stats := make(map[string]*types.ModelStats, len(fs.modelStats))
	for k, v := range fs.modelStats {
		stats[k] = v
	}

	return stats, nil
}

// ListModelStats retrieves all model statistics
func (fs *FileStorageImpl) ListModelStats() (map[string]*types.ModelStats, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	stats := make(map[string]*types.ModelStats, len(fs.modelStats))
	for k, v := range fs.modelStats {
		stats[k] = v
	}

	return stats, nil
}

// DeleteModelStats deletes model statistics
func (fs *FileStorageImpl) DeleteModelStats(model string) error {
	fs.mu.Lock()
	delete(fs.modelStats, model)
	fs.mu.Unlock()

	return fs.saveModelStats()
}

// SaveModelStatsHistory saves model statistics history
func (fs *FileStorageImpl) SaveModelStatsHistory(history *types.ModelStatsHistory) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Initialize history slice if not exists
	if _, exists := fs.modelHistory[history.Model]; !exists {
		fs.modelHistory[history.Model] = make([]*types.ModelStatsHistory, 0)
	}

	// Add new history entry
	fs.modelHistory[history.Model] = append(fs.modelHistory[history.Model], history)

	// Save to file
	historyFile := filepath.Join(fs.baseDir, fmt.Sprintf("%s_history.json", history.Model))
	data, err := json.MarshalIndent(fs.modelHistory[history.Model], "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(historyFile, data, 0644)
}

// GetModelStatsHistory retrieves model statistics history
func (fs *FileStorageImpl) GetModelStatsHistory(model string, limit int) ([]*types.ModelStatsHistory, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	history, exists := fs.modelHistory[model]
	if !exists {
		return nil, nil
	}

	if limit > 0 && len(history) > limit {
		return history[:limit], nil
	}

	return history, nil
}

// ListModelStatsHistory retrieves all model statistics history
func (fs *FileStorageImpl) ListModelStatsHistory(limit int) ([]*types.ModelStatsHistory, error) {
	fs.mu.RLock()
	defer fs.mu.RUnlock()

	var allHistory []*types.ModelStatsHistory
	for _, history := range fs.modelHistory {
		allHistory = append(allHistory, history...)
	}

	// Sort by timestamp in descending order
	sort.Slice(allHistory, func(i, j int) bool {
		return allHistory[i].Timestamp.After(allHistory[j].Timestamp)
	})

	if limit > 0 && len(allHistory) > limit {
		return allHistory[:limit], nil
	}

	return allHistory, nil
}

// DeleteModelStatsHistory deletes model statistics history
func (fs *FileStorageImpl) DeleteModelStatsHistory(model string) error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	delete(fs.modelHistory, model)
	return fs.saveModelHistory()
}

// Cleanup removes old data
func (fs *FileStorageImpl) Cleanup() error {
	fs.mu.Lock()
	defer fs.mu.Unlock()

	// Remove history entries older than 30 days
	thirtyDaysAgo := time.Now().Add(-30 * 24 * time.Hour)
	for model, history := range fs.modelHistory {
		var validHistory []*types.ModelStatsHistory
		for _, entry := range history {
			if entry.Timestamp.After(thirtyDaysAgo) {
				validHistory = append(validHistory, entry)
			}
		}
		fs.modelHistory[model] = validHistory
	}

	return fs.saveModelHistory()
}

// Close closes the storage
func (fs *FileStorageImpl) Close() error {
	return nil
}

// NewHistoryManager creates a new history manager
func (fs *FileStorageImpl) NewHistoryManager(size int) types.HistoryManager {
	return nil
}

// GetAllRequests 获取所有请求
func (fs *FileStorageImpl) GetAllRequests() ([]*types.Request, error) {
	// Not implemented
	return nil, nil
}

// GetRequests 获取指定用户的所有请求
func (fs *FileStorageImpl) GetRequests(userID string) ([]*types.Request, error) {
	// Not implemented
	return nil, nil
}

// GetRecentRequests 获取最近的请求记录
func (fs *FileStorageImpl) GetRecentRequests(limit int) ([]*types.Request, error) {
	// Not implemented
	return nil, nil
}

// GetRequestByID 根据ID获取请求
func (fs *FileStorageImpl) GetRequestByID(requestID string) (*types.Request, error) {
	return fs.GetRequest(requestID)
}
