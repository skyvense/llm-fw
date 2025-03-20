package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"sync"
	"time"

	"llm-fw/types"
)

// ModelInfo represents the model information with stats
type ModelInfo struct {
	Name        string                   `json:"name"`
	Family      string                   `json:"family,omitempty"`
	Parameters  string                   `json:"parameters,omitempty"`
	Format      string                   `json:"format,omitempty"`
	Stats       *types.ModelStats        `json:"stats,omitempty"`
	LastUsed    *time.Time               `json:"last_used,omitempty"`
	IsAvailable bool                     `json:"is_available"`
	History     *types.ModelStatsHistory `json:"history,omitempty"`
}

// OllamaModel used to parse Ollama API response
type OllamaModel struct {
	Name        string    `json:"name"`
	Model       string    `json:"model"`
	Modified_at time.Time `json:"modified_at"`
	Size        int64     `json:"size"`
	Digest      string    `json:"digest"`
	Details     struct {
		ParentModel       string   `json:"parent_model"`
		Format            string   `json:"format"`
		Family            string   `json:"family"`
		Families          []string `json:"families"`
		ParameterSize     string   `json:"parameter_size"`
		QuantizationLevel string   `json:"quantization_level"`
	} `json:"details"`
}

// OllamaResponse represents the full response from Ollama API
type OllamaResponse struct {
	Models []OllamaModel `json:"models"`
}

// ModelHandler handles model-related requests
type ModelHandler struct {
	ollamaURL        string
	storage          types.Storage
	metricsCollector types.MetricsCollector
	models           map[string]*types.ModelInfo
	mu               sync.RWMutex
}

// NewModelHandler creates a new model handler
func NewModelHandler(ollamaURL string, storage types.Storage, metricsCollector types.MetricsCollector) *ModelHandler {
	h := &ModelHandler{
		ollamaURL:        ollamaURL,
		storage:          storage,
		metricsCollector: metricsCollector,
		models:           make(map[string]*types.ModelInfo),
	}

	// 初始化时获取模型列表
	if err := h.refreshModels(); err != nil {
		log.Printf("Warning: failed to initialize model list: %v", err)
	}

	// 启动定期刷新
	go h.startRefreshLoop()
	return h
}

// refreshModels 从 Ollama 获取最新的模型列表
func (h *ModelHandler) refreshModels() error {
	log.Printf("Fetching models from Ollama at %s", h.ollamaURL)
	resp, err := http.Get(fmt.Sprintf("%s/api/tags", h.ollamaURL))
	if err != nil {
		return fmt.Errorf("failed to get models from Ollama: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %v", err)
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		log.Printf("Failed to parse Ollama response: %s", string(body))
		return fmt.Errorf("failed to parse response: %v", err)
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	// 更新模型列表
	for _, model := range ollamaResp.Models {
		modelName := strings.TrimSpace(model.Name)
		if _, exists := h.models[modelName]; !exists {
			log.Printf("Found new model: %s (Family: %s, Parameters: %s)",
				modelName, model.Details.Family, model.Details.ParameterSize)
			h.models[modelName] = &types.ModelInfo{
				Name:        modelName,
				Family:      model.Details.Family,
				Parameters:  model.Details.ParameterSize,
				Format:      model.Details.Format,
				IsAvailable: true,
			}
		} else {
			h.models[modelName].IsAvailable = true
		}
	}

	log.Printf("Successfully refreshed models, total count: %d", len(ollamaResp.Models))
	return nil
}

// startRefreshLoop 定期刷新模型列表
func (h *ModelHandler) startRefreshLoop() {
	log.Printf("Starting model refresh loop with 30-second interval")
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		log.Printf("Refreshing models...")
		if err := h.refreshModels(); err != nil {
			log.Printf("Failed to refresh models: %v", err)
		}
	}
}

// ListModels handles requests to get model list with their stats
func (h *ModelHandler) ListModels(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 获取查询参数
	statsOnly := r.URL.Query().Get("stats_only") == "true"

	// 获取所有模型的统计信息
	modelStats := h.metricsCollector.GetAllModelStats()

	// 获取模型历史数据
	history, err := h.storage.ListModelStatsHistory(10)
	if err != nil {
		log.Printf("Failed to get model history: %v", err)
		history = []*types.ModelStatsHistory{}
	}

	// 创建模型名称到历史数据的映射
	historyMap := make(map[string]*types.ModelStatsHistory)
	for _, entry := range history {
		historyMap[entry.Model] = entry
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	// 创建模型信息列表
	models := make([]*types.ModelInfo, 0)
	seenModels := make(map[string]bool)

	// 首先添加有统计信息的模型
	for modelName, stats := range modelStats {
		// 如果是 stats_only 模式，只添加有请求记录的模型
		if statsOnly && stats.TotalRequests == 0 {
			continue
		}

		modelName = strings.TrimSpace(modelName)
		seenModels[modelName] = true

		modelInfo, exists := h.models[modelName]
		if !exists {
			modelInfo = &types.ModelInfo{
				Name:        modelName,
				IsAvailable: false,
			}
		}
		modelInfo.Stats = stats
		modelInfo.LastUsed = &stats.LastUsed
		modelInfo.History = historyMap[modelName]

		models = append(models, modelInfo)
	}

	// 如果不是只显示有统计的模型，添加其他可用模型
	if !statsOnly {
		for modelName, modelInfo := range h.models {
			if !seenModels[modelName] {
				// 检查是否有历史数据
				if historyEntry, ok := historyMap[modelName]; ok {
					modelInfo.History = historyEntry
				}
				models = append(models, modelInfo)
			}
		}
	}

	// 按最后使用时间排序
	sort.Slice(models, func(i, j int) bool {
		if models[i].LastUsed == nil && models[j].LastUsed == nil {
			return models[i].Name < models[j].Name
		}
		if models[i].LastUsed == nil {
			return false
		}
		if models[j].LastUsed == nil {
			return true
		}
		return models[i].LastUsed.After(*models[j].LastUsed)
	})

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"models": models,
	})
}

// getModelNames 返回所有可用模型的名称列表
func getModelNames(stats map[string]*types.ModelStats) []string {
	var names []string
	for name := range stats {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// getModelsFromOllama retrieves the list of models from Ollama
func (h *ModelHandler) getModelsFromOllama() ([]types.ModelInfo, error) {
	// 发送请求到 Ollama API
	resp, err := http.Get(h.ollamaURL + "/api/tags")
	if err != nil {
		return nil, fmt.Errorf("failed to get models from Ollama: %v", err)
	}
	defer resp.Body.Close()

	// 读取响应
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %v", err)
	}

	// 解析响应
	var ollamaResp OllamaResponse
	if err := json.Unmarshal(body, &ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to parse response: %v", err)
	}

	// 转换为 ModelInfo 列表
	var models []types.ModelInfo
	h.mu.Lock()
	defer h.mu.Unlock()

	for _, m := range ollamaResp.Models {
		modelInfo := types.ModelInfo{
			Name:        m.Name,
			Family:      m.Details.Family,
			Parameters:  m.Details.ParameterSize,
			Format:      m.Details.Format,
			IsAvailable: true,
		}
		models = append(models, modelInfo)
		// 更新内部模型列表
		h.models[m.Name] = &modelInfo
	}

	return models, nil
}

// GetTags 提供兼容 Ollama API 的 /api/tags 接口
func (h *ModelHandler) GetTags(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	// 转换为 Ollama API 格式
	ollamaModels := make([]OllamaModel, 0)
	for _, model := range h.models {
		if !model.IsAvailable {
			continue
		}

		ollamaModel := OllamaModel{
			Name:        model.Name,
			Model:       model.Name,
			Modified_at: time.Now(), // 由于我们不跟踪修改时间，使用当前时间
			Size:        0,          // 我们不跟踪模型大小
			Details: struct {
				ParentModel       string   `json:"parent_model"`
				Format            string   `json:"format"`
				Family            string   `json:"family"`
				Families          []string `json:"families"`
				ParameterSize     string   `json:"parameter_size"`
				QuantizationLevel string   `json:"quantization_level"`
			}{
				Format:        model.Format,
				Family:        model.Family,
				Families:      []string{model.Family},
				ParameterSize: model.Parameters,
			},
		}
		ollamaModels = append(ollamaModels, ollamaModel)
	}

	// 返回响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(OllamaResponse{
		Models: ollamaModels,
	})
}
