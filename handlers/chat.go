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

// ChatRequest 定义了聊天请求的结构
type ChatRequest struct {
	Model    string        `json:"model" binding:"required"`
	Messages []ChatMessage `json:"messages" binding:"required"`
	UserID   string        `json:"user_id"`
	Stream   bool          `json:"stream"`
}

// ChatMessage 定义了聊天消息的结构
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatHandler 处理聊天相关的请求
type ChatHandler struct {
	TargetURL        string
	Storage          Storage
	MetricsCollector MetricsCollector
}

// NewChatHandler 创建一个新的聊天处理器
func NewChatHandler(targetURL string, storage Storage, metricsCollector MetricsCollector) *ChatHandler {
	return &ChatHandler{
		TargetURL:        targetURL,
		Storage:          storage,
		MetricsCollector: metricsCollector,
	}
}

// Chat 处理聊天请求
func (h *ChatHandler) Chat(c *gin.Context) {
	var req ChatRequest
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

	// 设置 stream 为 true 以获取 token 信息
	req.Stream = true
	jsonBody, err := json.Marshal(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to build request"})
		return
	}

	proxyReq, err := http.NewRequest("POST", h.TargetURL+"/api/chat", bytes.NewBuffer(jsonBody))
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
			int64(len(req.Messages[len(req.Messages)-1].Content)),
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
			int64(len(req.Messages[len(req.Messages)-1].Content)),
			0,
			time.Since(startTime).Milliseconds(),
			false,
		)
		c.JSON(proxyResp.StatusCode, gin.H{"error": "model server returned error"})
		return
	}

	// 设置响应头，支持流式传输
	c.Header("Content-Type", "application/x-ndjson")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	// 读取流式响应
	decoder := json.NewDecoder(proxyResp.Body)
	var fullResponse strings.Builder
	var totalTokensIn int64
	var totalTokensOut int64
	var lastError error

	for {
		var streamResp struct {
			Message struct {
				Content string `json:"content"`
				Role    string `json:"role"`
			} `json:"message"`
			Done   bool `json:"done"`
			Prompt struct {
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

		// 更新 token 计数
		if len(streamResp.Prompt.Tokens) > 0 {
			totalTokensIn = int64(len(streamResp.Prompt.Tokens))
		}
		if len(streamResp.Generation.Tokens) > 0 {
			totalTokensOut += int64(len(streamResp.Generation.Tokens))
		}

		// 构建流式响应
		response := gin.H{
			"message": streamResp.Message,
			"done":    streamResp.Done,
		}

		// 如果是最后一个响应，添加统计信息
		if streamResp.Done {
			response["stats"] = gin.H{
				"tokens_in":  totalTokensIn,
				"tokens_out": totalTokensOut,
				"latency_ms": time.Since(startTime).Milliseconds(),
			}
		}

		// 发送流式响应
		data, err := json.Marshal(response)
		if err != nil {
			lastError = err
			break
		}
		c.Writer.Write(data)
		c.Writer.Write([]byte("\n"))
		c.Writer.Flush()

		fullResponse.WriteString(streamResp.Message.Content)

		if streamResp.Done {
			break
		}
	}

	if lastError != nil {
		log.Printf("Error during streaming: %v", lastError)
		h.MetricsCollector.RecordRequest(
			req.Model,
			"proxy",
			totalTokensIn,
			totalTokensOut,
			time.Since(startTime).Milliseconds(),
			false,
		)
		return
	}

	response := fullResponse.String()
	latency := time.Since(startTime).Milliseconds()

	// 如果没有获取到 token 数量，使用估算值
	if totalTokensIn == 0 {
		var totalPromptLength int64
		for _, msg := range req.Messages {
			totalPromptLength += int64(len(msg.Content))
		}
		totalTokensIn = totalPromptLength / 4
	}
	if totalTokensOut == 0 {
		totalTokensOut = int64(len(response) / 4)
	}

	// 规范化模型名称
	modelName := strings.TrimSpace(req.Model)
	log.Printf("Chat request details:")
	log.Printf("  Model: %s", modelName)
	log.Printf("  User: %s", req.UserID)
	log.Printf("  Messages count: %d", len(req.Messages))
	log.Printf("  Last message length: %d characters", len(req.Messages[len(req.Messages)-1].Content))
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
			log.Printf("Model statistics after chat request:")
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
		Model:     modelName,
		Prompt:    req.Messages[len(req.Messages)-1].Content,
		Response:  response,
		TokensIn:  int(totalTokensIn),
		TokensOut: int(totalTokensOut),
		Server:    "proxy",
	}

	if err := h.Storage.SaveRequest(storageReq); err != nil {
		log.Printf("Failed to save chat request: %v", err)
	}
}
