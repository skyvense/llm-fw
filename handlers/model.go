package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"strings"
	"time"

	"llm-fw/metrics"

	"github.com/gin-gonic/gin"
)

// OllamaModel used to parse Ollama API response
type OllamaModel struct {
	Name        string    `json:"name"`
	Modified_at time.Time `json:"modified_at"`
	Size        int64     `json:"size"`
	Digest      string    `json:"digest"`
}

// ModelHandler handles model-related requests
type ModelHandler struct {
	TargetURL        string
	MetricsCollector MetricsCollector
}

// NewModelHandler creates a new model handler
func NewModelHandler(targetURL string, metricsCollector MetricsCollector) *ModelHandler {
	return &ModelHandler{
		TargetURL:        targetURL,
		MetricsCollector: metricsCollector,
	}
}

// ListModels handles requests to get model list
func (h *ModelHandler) ListModels(c *gin.Context) {
	startTime := time.Now()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	apiURL := h.TargetURL + "/api/tags"
	log.Printf("Attempting to fetch model list: %s", apiURL)

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to create request: %v", err),
		})
		return
	}

	req.Header.Set("Accept", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to send request: %v", err),
		})
		return
	}
	defer resp.Body.Close()

	log.Printf("Received response status code: %d", resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("server returned error status code: %d", resp.StatusCode),
		})
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to read response: %v", err),
		})
		return
	}

	log.Printf("Received response content: %s", string(body))

	var response struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("failed to parse response: %v", err),
		})
		return
	}

	var modelList []string
	for _, model := range response.Models {
		modelList = append(modelList, model.Name)
	}

	if len(modelList) == 0 {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "no models found",
		})
		return
	}

	sort.Strings(modelList)
	latency := time.Since(startTime).Milliseconds()

	h.MetricsCollector.RecordRequest(
		"system",
		"list_models",
		0,
		int64(len(modelList)),
		latency,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"models":     modelList,
		"total":      len(modelList),
		"latency_ms": latency,
	})
}

// GetModelStats handles requests to get model statistics
func (h *ModelHandler) GetModelStats(c *gin.Context) {
	model := strings.TrimSpace(c.Param("model"))
	log.Printf("Getting stats for model: %s", model)

	metrics := h.MetricsCollector.GetMetrics()
	if metrics == nil {
		log.Printf("No metrics data available")
		c.JSON(http.StatusOK, gin.H{
			"total_requests":  0,
			"total_tokens":    0,
			"average_latency": 0,
			"success_rate":    0,
		})
		return
	}

	modelNames := getModelNames(metrics.ModelStats)
	log.Printf("Available models in stats: %v", modelNames)
	log.Printf("Looking for exact match of model: %s", model)

	if stats, exists := metrics.ModelStats[model]; exists {
		log.Printf("Found stats for model %s: %+v", model, stats)
		totalTokens := stats.TotalTokensIn + stats.TotalTokensOut
		successRate := 0.0
		if stats.TotalRequests > 0 {
			successRate = float64(stats.TotalRequests-stats.FailedRequests) / float64(stats.TotalRequests) * 100
		}

		c.JSON(http.StatusOK, gin.H{
			"total_requests":  stats.TotalRequests,
			"total_tokens":    totalTokens,
			"average_latency": stats.AverageLatency,
			"success_rate":    successRate,
		})
	} else {
		log.Printf("No stats found for model: %s (available models: %v)", model, modelNames)
		c.JSON(http.StatusOK, gin.H{
			"total_requests":  0,
			"total_tokens":    0,
			"average_latency": 0,
			"success_rate":    0,
		})
	}
}

// getModelNames 返回所有可用模型的名称列表
func getModelNames(stats map[string]*metrics.ModelMetrics) []string {
	var names []string
	for name := range stats {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
