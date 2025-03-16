package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GenerateRequest 定义了生成请求的结构
type GenerateRequest struct {
	Model  string `json:"model" binding:"required"`
	Prompt string `json:"prompt" binding:"required"`
	UserID string `json:"user_id"`
}

// GenerateHandler 处理生成相关的请求
type GenerateHandler struct {
	TargetURL        string
	Storage          Storage
	MetricsCollector MetricsCollector
}

// NewGenerateHandler 创建一个新的生成处理器
func NewGenerateHandler(targetURL string, storage Storage, metricsCollector MetricsCollector) *GenerateHandler {
	return &GenerateHandler{
		TargetURL:        targetURL,
		Storage:          storage,
		MetricsCollector: metricsCollector,
	}
}

// Generate 处理生成请求
func (h *GenerateHandler) Generate(c *gin.Context) {
	var req GenerateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.UserID == "" {
		req.UserID = "anonymous_" + uuid.New().String()[:8]
	}

	startTime := time.Now()

	client := &http.Client{
		Timeout: 300 * time.Second,
	}

	requestBody := map[string]interface{}{
		"model":  req.Model,
		"prompt": req.Prompt,
		"stream": true,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build request"})
		return
	}

	proxyReq, err := http.NewRequest("POST", h.TargetURL+"/api/generate", bytes.NewBuffer(jsonBody))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
		return
	}
	proxyReq.Header.Set("Content-Type", "application/json")

	proxyResp, err := client.Do(proxyReq)
	if err != nil {
		log.Printf("Failed to send request to model server: %v", err)
		h.MetricsCollector.RecordRequest(
			req.Model,
			"proxy",
			int64(len(req.Prompt)),
			0,
			time.Since(startTime).Milliseconds(),
			false,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer proxyResp.Body.Close()

	if proxyResp.StatusCode != http.StatusOK {
		log.Printf("Model server returned error status code: %d", proxyResp.StatusCode)
		h.MetricsCollector.RecordRequest(
			req.Model,
			"proxy",
			int64(len(req.Prompt)),
			0,
			time.Since(startTime).Milliseconds(),
			false,
		)
		c.JSON(proxyResp.StatusCode, gin.H{"error": "model server returned error"})
		return
	}

	// 读取流式响应
	decoder := json.NewDecoder(proxyResp.Body)
	var fullResponse strings.Builder
	var lastError error
	var totalTokensIn int64
	var totalTokensOut int64

	for {
		var streamResp struct {
			Response string `json:"response"`
			Done     bool   `json:"done"`
			Prompt   struct {
				Tokens []any `json:"tokens"`
			} `json:"prompt"`
			Generation struct {
				Tokens []any `json:"tokens"`
			} `json:"generation"`
		}

		if err := decoder.Decode(&streamResp); err != nil {
			if err == io.EOF {
				break
			}
			lastError = err
			break
		}

		fullResponse.WriteString(streamResp.Response)

		// 累计 token 数量
		if len(streamResp.Prompt.Tokens) > 0 {
			totalTokensIn = int64(len(streamResp.Prompt.Tokens))
		}
		if len(streamResp.Generation.Tokens) > 0 {
			totalTokensOut += int64(len(streamResp.Generation.Tokens))
		}

		if streamResp.Done {
			break
		}
	}

	if lastError != nil {
		log.Printf("Failed to parse response: %v", lastError)
		h.MetricsCollector.RecordRequest(
			req.Model,
			"proxy",
			totalTokensIn,
			totalTokensOut,
			time.Since(startTime).Milliseconds(),
			false,
		)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse response"})
		return
	}

	response := fullResponse.String()
	latency := time.Since(startTime).Milliseconds()

	// 如果没有获取到 token 数量，使用估算值
	if totalTokensIn == 0 {
		totalTokensIn = int64(len(req.Prompt) / 4) // 粗略估算，每个 token 约 4 个字符
	}
	if totalTokensOut == 0 {
		totalTokensOut = int64(len(response) / 4)
	}

	// 规范化模型名称
	modelName := strings.TrimSpace(req.Model)
	log.Printf("Request details:")
	log.Printf("  Model: %s", modelName)
	log.Printf("  User: %s", req.UserID)
	log.Printf("  Prompt length: %d characters", len(req.Prompt))
	log.Printf("  Response length: %d characters", len(response))
	log.Printf("  Tokens In: %d", totalTokensIn)
	log.Printf("  Tokens Out: %d", totalTokensOut)
	log.Printf("  Latency: %dms", latency)
	log.Printf("  Server: proxy")

	// 记录请求统计
	h.MetricsCollector.RecordRequest(
		modelName,
		"proxy",
		totalTokensIn,
		totalTokensOut,
		latency,
		true,
	)

	// 验证统计信息是否已记录
	metrics := h.MetricsCollector.GetMetrics()
	if metrics != nil {
		if stats, exists := metrics.ModelStats[modelName]; exists {
			log.Printf("Model statistics after request:")
			log.Printf("  Total requests: %d", stats.TotalRequests)
			log.Printf("  Total tokens in: %d", stats.TotalTokensIn)
			log.Printf("  Total tokens out: %d", stats.TotalTokensOut)
			log.Printf("  Average latency: %.2fms", stats.AverageLatency)
			log.Printf("  Failed requests: %d", stats.FailedRequests)
		} else {
			log.Printf("Warning: Stats not found for model %s immediately after recording", modelName)
		}
	}

	storageReq := &Request{
		ID:        uuid.New().String(),
		UserID:    req.UserID,
		Model:     modelName, // 使用规范化的模型名称
		Prompt:    req.Prompt,
		Response:  response,
		TokensIn:  int(totalTokensIn),
		TokensOut: int(totalTokensOut),
		Server:    "proxy",
	}

	if err := h.Storage.SaveRequest(storageReq); err != nil {
		log.Printf("Failed to save request: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save request record"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"response": response,
		"user_id":  req.UserID,
		"stats": gin.H{
			"tokens_in":  totalTokensIn,
			"tokens_out": totalTokensOut,
			"latency_ms": latency,
		},
	})
}
