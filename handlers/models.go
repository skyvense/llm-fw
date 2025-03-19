package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"llm-fw/types"
)

// ModelsHandler handles model-related requests
type ModelsHandler struct {
	storage types.Storage
}

// NewModelsHandler creates a new ModelsHandler
func NewModelsHandler(storage types.Storage) *ModelsHandler {
	return &ModelsHandler{storage: storage}
}

// HandleGetModels handles GET /api/models requests
func (h *ModelsHandler) HandleGetModels(w http.ResponseWriter, r *http.Request) {
	// Get query parameters
	statsOnly := r.URL.Query().Get("stats_only") == "true"

	// Get models from Ollama
	models, err := h.getModelsFromOllama()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get models from Ollama: %v", err), http.StatusInternalServerError)
		return
	}

	// If stats_only is true, return only models with stats
	if statsOnly {
		modelsWithStats := make([]types.ModelInfo, 0)
		for _, model := range models {
			stats, err := h.storage.GetModelStats(model.Name)
			if err != nil {
				continue
			}
			model.Stats = stats
			modelsWithStats = append(modelsWithStats, model)
		}
		models = modelsWithStats
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"models": models,
	})
}

// HandleGetModelStats handles GET /api/models/{model}/stats requests
func (h *ModelsHandler) HandleGetModelStats(w http.ResponseWriter, r *http.Request) {
	model := r.URL.Query().Get("model")

	if model == "" {
		http.Error(w, "Model name is required", http.StatusBadRequest)
		return
	}

	// Get model stats from storage
	stats, err := h.storage.GetModelStats(model)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get model stats: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"model": model,
		"stats": stats,
	})
}

// HandleGetModelHistory handles GET /api/models/{model}/history requests
func (h *ModelsHandler) HandleGetModelHistory(w http.ResponseWriter, r *http.Request) {
	model := r.URL.Query().Get("model")
	limit := 100 // Default limit

	// Parse limit parameter
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if _, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil {
			http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
			return
		}
	}

	if model == "" {
		http.Error(w, "Model name is required", http.StatusBadRequest)
		return
	}

	// Get model history from storage
	history, err := h.storage.GetModelStatsHistory(model, limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get model history: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"model":   model,
		"history": history,
	})
}

// HandleGetAllModelHistory handles GET /api/models/history requests
func (h *ModelsHandler) HandleGetAllModelHistory(w http.ResponseWriter, r *http.Request) {
	limit := 100 // Default limit

	// Parse limit parameter
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if _, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil {
			http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
			return
		}
	}

	// Get all model history from storage
	history, err := h.storage.ListModelStatsHistory(limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get model history: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"history": history,
	})
}

// getModelsFromOllama retrieves the list of models from Ollama
func (h *ModelsHandler) getModelsFromOllama() ([]types.ModelInfo, error) {
	// TODO: Implement Ollama API call to get models
	// For now, return a mock response
	return []types.ModelInfo{
		{
			Name:        "deepseek-r1:1.5b",
			IsAvailable: true,
		},
		{
			Name:        "deepseek-r1:7b",
			IsAvailable: true,
		},
	}, nil
}

// UpdateModelStats updates the statistics for a model
func (h *ModelsHandler) UpdateModelStats(model string, stats *types.ModelStats) error {
	// Save current stats
	if err := h.storage.SaveModelStats(model, stats); err != nil {
		return fmt.Errorf("failed to save model stats: %w", err)
	}

	// Create and save history entry
	history := &types.ModelStatsHistory{
		ID:             fmt.Sprintf("%s-%d", model, time.Now().UnixNano()),
		Model:          model,
		TotalRequests:  stats.TotalRequests,
		FailedRequests: stats.FailedRequests,
		TotalTokensIn:  stats.TotalTokensIn,
		TotalTokensOut: stats.TotalTokensOut,
		AverageLatency: stats.AverageLatency,
		Timestamp:      time.Now(),
	}

	if err := h.storage.SaveModelStatsHistory(history); err != nil {
		return fmt.Errorf("failed to save model history: %w", err)
	}

	return nil
}
