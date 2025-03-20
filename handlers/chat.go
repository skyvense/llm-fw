package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"llm-fw/types"
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

// ChatResponse 定义聊天响应的结构
type ChatResponse struct {
	Response string             `json:"response"`
	Stats    types.RequestStats `json:"stats"`
}

// ChatHandler 处理聊天相关的请求
type ChatHandler struct {
	TargetURL        string
	Storage          types.Storage
	MetricsCollector types.MetricsCollector
	ollamaURL        string
}

// NewChatHandler creates a new chat handler
func NewChatHandler(storage types.Storage, ollamaURL string, metricsCollector types.MetricsCollector) *ChatHandler {
	return &ChatHandler{
		Storage:          storage,
		ollamaURL:        ollamaURL,
		TargetURL:        ollamaURL,
		MetricsCollector: metricsCollector,
	}
}

// Chat 处理聊天请求
func (h *ChatHandler) Chat(c *gin.Context) {
	if c.Request.Method != http.MethodPost {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method not allowed"})
		return
	}

	// 设置 CORS 头
	c.Header("Access-Control-Allow-Origin", "http://127.0.0.1")
	c.Header("Access-Control-Allow-Methods", "POST, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type")
	c.Header("Vary", "Origin")

	// 处理 OPTIONS 请求
	if c.Request.Method == http.MethodOptions {
		c.Status(http.StatusOK)
		return
	}

	var req ChatRequest
	if err := json.NewDecoder(c.Request.Body).Decode(&req); err != nil {
		// 如果 JSON 解析失败，尝试从查询参数获取
		model := c.Query("model")
		message := c.Query("message")
		if model != "" && message != "" {
			req.Model = model
			req.Messages = []ChatMessage{
				{
					Role:    "user",
					Content: message,
				},
			}
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request body or missing parameters"})
			return
		}
	}

	// 验证请求
	if req.Model == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Model is required"})
		return
	}
	if len(req.Messages) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "At least one message is required"})
		return
	}

	// 从请求头中获取用户ID，如果没有则生成一个
	userID := c.GetHeader("X-User-ID")
	if userID == "" {
		userID = "anonymous_" + uuid.New().String()[:8]
	}
	req.UserID = userID

	startTime := time.Now()

	// 调用Ollama API
	ollamaReq := map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
		"stream":   true, // 始终启用流式响应
	}

	ollamaReqBody, err := json.Marshal(ollamaReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to marshal request"})
		return
	}

	resp, err := http.Post(fmt.Sprintf("%s/api/chat", h.TargetURL), "application/json", bytes.NewBuffer(ollamaReqBody))
	if err != nil {
		log.Printf("Failed to call Ollama API: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to call Ollama API"})
		return
	}
	defer resp.Body.Close()

	// 设置响应头
	c.Header("Content-Type", "application/x-ndjson")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	// 创建响应通道
	responseChan := make(chan string, 100)
	errorChan := make(chan error, 1)

	// 创建共享变量
	var fullResponse strings.Builder
	var promptEvalCount, evalCount float64

	// 启动goroutine处理响应
	go func() {
		decoder := json.NewDecoder(resp.Body)

		for decoder.More() {
			var chunk map[string]interface{}
			if err := decoder.Decode(&chunk); err != nil {
				errorChan <- fmt.Errorf("failed to decode response chunk: %v", err)
				return
			}

			// 从消息中提取响应文本
			if message, ok := chunk["message"].(map[string]interface{}); ok {
				if content, ok := message["content"].(string); ok {
					fullResponse.WriteString(content)
					responseChan <- content
				}
			}

			// 更新token计数
			if count, ok := chunk["prompt_eval_count"].(float64); ok {
				promptEvalCount = count
			}
			if count, ok := chunk["eval_count"].(float64); ok {
				evalCount = count
			}
		}

		// 发送完成信号
		responseChan <- ""
	}()

	// 处理响应流
	for {
		select {
		case content := <-responseChan:
			if content == "" {
				// 响应完成
				latency := time.Since(startTime).Milliseconds()
				stats := types.RequestStats{
					TokensIn:  int(promptEvalCount),
					TokensOut: int(evalCount),
					LatencyMs: float64(latency),
				}

				// 更新指标
				h.MetricsCollector.RecordRequest(&types.Request{
					Model:     req.Model,
					Server:    "ollama",
					TokensIn:  int(stats.TokensIn),
					TokensOut: int(stats.TokensOut),
					LatencyMs: float64(latency),
					Status:    0,
					Timestamp: time.Now(),
				})

				// 保存到存储
				storageReq := &types.Request{
					ID:        uuid.New().String(),
					UserID:    req.UserID,
					Model:     req.Model,
					Prompt:    req.Messages[len(req.Messages)-1].Content,
					Response:  fullResponse.String(),
					TokensIn:  stats.TokensIn,
					TokensOut: stats.TokensOut,
					Server:    "ollama",
					LatencyMs: float64(latency),
					Timestamp: time.Now(),
					Source:    "external_ui", // 标记请求来源
				}

				if err := h.Storage.SaveRequest(storageReq); err != nil {
					log.Printf("Failed to save chat request: %v", err)
				}

				// 发送最终统计信息
				finalResponse := map[string]interface{}{
					"model":      req.Model,
					"created_at": time.Now().Format(time.RFC3339),
					"message": map[string]interface{}{
						"role":    "assistant",
						"content": fullResponse.String(),
					},
					"done": true,
					"stats": map[string]interface{}{
						"prompt_eval_count": promptEvalCount,
						"eval_count":        evalCount,
						"eval_duration":     float64(latency) / 1000.0,
					},
				}
				finalJSON, _ := json.Marshal(finalResponse)
				c.Writer.Write(finalJSON)
				c.Writer.Write([]byte("\n"))
				c.Writer.Flush()
				return
			}
			// 发送内容块
			messageEvent := map[string]interface{}{
				"model":      req.Model,
				"created_at": time.Now().Format(time.RFC3339),
				"message": map[string]interface{}{
					"role":    "assistant",
					"content": content,
				},
			}
			jsonData, _ := json.Marshal(messageEvent)
			c.Writer.Write(jsonData)
			c.Writer.Write([]byte("\n"))
			c.Writer.Flush()
		case err := <-errorChan:
			log.Printf("Error processing response: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process response"})
			return
		}
	}
}

// HandleGetHistory handles GET /api/history requests
func (h *ChatHandler) HandleGetHistory(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	limit := 100 // Default limit

	// Parse limit parameter
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if _, err := fmt.Sscanf(limitStr, "%d", &limit); err != nil {
			http.Error(w, "Invalid limit parameter", http.StatusBadRequest)
			return
		}
	}

	// Get history from storage
	requests, err := h.Storage.ListRequests(limit)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get history: %v", err), http.StatusInternalServerError)
		return
	}

	// Return response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"requests": requests,
	})
}

// HandleChat handles chat requests
func (h *ChatHandler) HandleChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// ... rest of the implementation ...
}
