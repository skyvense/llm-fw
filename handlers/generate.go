package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
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

	client := &http.Client{}

	requestBody := map[string]interface{}{
		"model":  req.Model,
		"prompt": req.Prompt,
	}
	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "构建请求失败"})
		return
	}

	proxyReq, err := http.NewRequest("POST", h.TargetURL+"/api/generate", bytes.NewBuffer(jsonBody))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "创建请求失败"})
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

	var ollamaResp struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(proxyResp.Body).Decode(&ollamaResp); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "解析响应失败"})
		return
	}

	latency := time.Since(startTime).Milliseconds()

	storageReq := &Request{
		ID:        uuid.New().String(),
		UserID:    req.UserID,
		Model:     req.Model,
		Prompt:    req.Prompt,
		Response:  ollamaResp.Response,
		TokensIn:  len(req.Prompt),
		TokensOut: len(ollamaResp.Response),
		Server:    "proxy",
	}

	if err := h.Storage.SaveRequest(storageReq); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存请求记录失败"})
		return
	}

	h.MetricsCollector.RecordRequest(
		req.Model,
		"proxy",
		int64(len(req.Prompt)),
		int64(len(ollamaResp.Response)),
		latency,
		true,
	)

	c.JSON(http.StatusOK, gin.H{
		"response": ollamaResp.Response,
		"user_id":  req.UserID,
	})
}
