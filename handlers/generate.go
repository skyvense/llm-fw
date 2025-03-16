package handlers

import (
	"bytes"
	"encoding/json"
	"io"
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
		"stream": true, // 启用流式响应
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
		c.JSON(proxyResp.StatusCode, gin.H{"error": "model server returned error"})
		return
	}

	// 读取流式响应
	decoder := json.NewDecoder(proxyResp.Body)
	var fullResponse strings.Builder
	var lastError error

	for {
		var streamResp struct {
			Response string `json:"response"`
			Done     bool   `json:"done"`
		}

		if err := decoder.Decode(&streamResp); err != nil {
			if err == io.EOF {
				break
			}
			lastError = err
			break
		}

		fullResponse.WriteString(streamResp.Response)

		if streamResp.Done {
			break
		}
	}

	if lastError != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to parse response"})
		return
	}

	response := fullResponse.String()

	latency := time.Since(startTime).Milliseconds()

	storageReq := &Request{
		ID:        uuid.New().String(),
		UserID:    req.UserID,
		Model:     req.Model,
		Prompt:    req.Prompt,
		Response:  response,
		TokensIn:  len(req.Prompt),
		TokensOut: len(response),
		Server:    "proxy",
	}

	if err := h.Storage.SaveRequest(storageReq); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save request record"})
		return
	}

	h.MetricsCollector.RecordRequest(
		req.Model,
		"proxy",
		int64(len(req.Prompt)),
		int64(len(response)),
		latency,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"response": response,
		"user_id":  req.UserID,
	})
}
