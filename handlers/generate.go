package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"llm-fw/api"
	"llm-fw/interfaces"
)

// GenerateRequest 定义了生成请求的结构
type GenerateRequest struct {
	Model  string `json:"model" binding:"required"`
	Prompt string `json:"prompt" binding:"required"`
	UserID string `json:"user_id"`
	Stream bool   `json:"stream"`
}

// GenerateResponse 定义生成响应的结构
type GenerateResponse struct {
	Response string           `json:"response"`
	Stats    api.RequestStats `json:"stats"`
}

// GenerateHandler 处理生成相关的请求
type GenerateHandler struct {
	TargetURL        string
	Storage          interfaces.Storage
	MetricsCollector interfaces.MetricsCollector
}

// NewGenerateHandler 创建一个新的生成处理器
func NewGenerateHandler(targetURL string, storage interfaces.Storage, metricsCollector interfaces.MetricsCollector) *GenerateHandler {
	return &GenerateHandler{
		TargetURL:        targetURL,
		Storage:          storage,
		MetricsCollector: metricsCollector,
	}
}

// Generate 处理生成请求
func (h *GenerateHandler) Generate(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method not allowed"})
		return
	}

	var req GenerateRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body"})
		return
	}

	if req.UserID == "" {
		req.UserID = "anonymous_" + uuid.New().String()[:8]
	}

	startTime := time.Now()

	// 调用Ollama API
	ollamaReq := map[string]interface{}{
		"model":  req.Model,
		"prompt": req.Prompt,
	}

	ollamaReqBody, err := json.Marshal(ollamaReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal request"})
		return
	}

	resp, err := http.Post(fmt.Sprintf("%s/api/generate", h.TargetURL), "application/json", bytes.NewBuffer(ollamaReqBody))
	if err != nil {
		log.Printf("Failed to call Ollama API: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call Ollama API"})
		return
	}
	defer resp.Body.Close()

	// 读取流式响应
	decoder := json.NewDecoder(resp.Body)
	var fullResponse string
	var promptEvalCount, evalCount float64

	for decoder.More() {
		var chunk map[string]interface{}
		if err := decoder.Decode(&chunk); err != nil {
			log.Printf("Failed to decode response chunk: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode response"})
			return
		}

		// 从响应中提取文本
		if response, ok := chunk["response"].(string); ok {
			fullResponse += response
		}

		// 更新token计数
		if count, ok := chunk["prompt_eval_count"].(float64); ok {
			promptEvalCount = count
		}
		if count, ok := chunk["eval_count"].(float64); ok {
			evalCount = count
		}
	}

	latency := time.Since(startTime).Milliseconds()

	// 创建统计信息
	stats := api.RequestStats{
		TokensIn:  int(promptEvalCount),
		TokensOut: int(evalCount),
		LatencyMs: float64(latency),
	}

	// 更新指标
	h.MetricsCollector.RecordRequest(
		req.Model,
		"ollama",
		int64(stats.TokensIn),
		int64(stats.TokensOut),
		latency,
		true,
	)

	// 保存到存储
	storageReq := &api.Request{
		ID:        uuid.New().String(),
		UserID:    req.UserID,
		Model:     req.Model,
		Prompt:    req.Prompt,
		Response:  fullResponse,
		TokensIn:  stats.TokensIn,
		TokensOut: stats.TokensOut,
		Server:    "ollama",
		Timestamp: time.Now(),
	}

	if err := h.Storage.SaveRequest(storageReq); err != nil {
		log.Printf("Failed to save generate request: %v", err)
	}

	// 返回响应
	c.Header("Content-Type", "application/json")
	c.JSON(http.StatusOK, GenerateResponse{
		Response: fullResponse,
		Stats:    stats,
	})
}
