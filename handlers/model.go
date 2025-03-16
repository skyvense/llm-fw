package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"sort"
	"time"

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
	model := c.Param("model")
	log.Printf("Getting stats for model: %s", model)

	metrics := h.MetricsCollector.GetMetrics()
	if metrics == nil {
		log.Printf("No metrics data available")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "metrics data not available",
		})
		return
	}

	log.Printf("Available model stats: %+v", metrics.ModelStats)

	modelStats := struct {
		TotalRequests  int64   `json:"total_requests"`
		TotalTokens    int64   `json:"total_tokens"`
		AverageLatency float64 `json:"average_latency"`
		SuccessRate    float64 `json:"success_rate"`
	}{}

	if stats, exists := metrics.ModelStats[model]; exists {
		log.Printf("Found stats for model %s: %+v", model, stats)
		modelStats.TotalRequests = stats.TotalRequests
		modelStats.TotalTokens = stats.TotalTokensIn + stats.TotalTokensOut
		if stats.TotalRequests > 0 {
			modelStats.AverageLatency = float64(stats.TotalLatency) / float64(stats.TotalRequests)
			modelStats.SuccessRate = float64(stats.TotalRequests-stats.FailedRequests) / float64(stats.TotalRequests) * 100
		}
		c.JSON(http.StatusOK, modelStats)
	} else {
		log.Printf("No stats found for model: %s", model)
		// 初始化一个空的统计信息而不是返回错误
		c.JSON(http.StatusOK, modelStats)
	}
}
