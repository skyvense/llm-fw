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

	"llm-fw/types"
)

// GenerateRequest 定义了生成请求的结构
type GenerateRequest struct {
	Model       string   `json:"model" binding:"required"`
	Prompt      string   `json:"prompt" binding:"required"`
	MaxTokens   int      `json:"max_tokens,omitempty"`
	Temperature float64  `json:"temperature,omitempty"`
	TopP        float64  `json:"top_p,omitempty"`
	N           int      `json:"n,omitempty"`
	Stream      bool     `json:"stream,omitempty"`
	Stop        []string `json:"stop,omitempty"`
}

// GenerateResponse 定义了生成响应的结构
type GenerateResponse struct {
	ID      string   `json:"id"`
	Object  string   `json:"object"`
	Created int64    `json:"created"`
	Model   string   `json:"model"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice 定义了选择的结构
type Choice struct {
	Text         string      `json:"text"`
	Index        int         `json:"index"`
	LogProbs     interface{} `json:"logprobs"`
	FinishReason string      `json:"finish_reason"`
}

// Usage 定义了使用统计的结构
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// GenerateHandler 处理生成相关的请求
type GenerateHandler struct {
	TargetURL        string
	Storage          types.Storage
	MetricsCollector types.MetricsCollector
}

// NewGenerateHandler 创建一个新的生成处理器
func NewGenerateHandler(targetURL string, storage types.Storage, metricsCollector types.MetricsCollector) *GenerateHandler {
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

	// 设置默认值
	if req.N == 0 {
		req.N = 1
	}
	if req.MaxTokens == 0 {
		req.MaxTokens = 2048
	}
	if req.Temperature == 0 {
		req.Temperature = 0.7
	}
	if req.TopP == 0 {
		req.TopP = 1
	}

	startTime := time.Now()

	// 调用Ollama API
	ollamaReq := map[string]interface{}{
		"model":  req.Model,
		"prompt": req.Prompt,
		"stream": req.Stream,
		"options": map[string]interface{}{
			"num_predict": req.MaxTokens,
			"temperature": req.Temperature,
			"top_p":       req.TopP,
			"stop":        req.Stop,
		},
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

	// 创建响应
	response := GenerateResponse{
		ID:      uuid.New().String(),
		Object:  "text_completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
		Choices: make([]Choice, req.N),
	}

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
		if responseText, ok := chunk["response"].(string); ok {
			fullResponse += responseText
			if req.Stream {
				// 发送流式响应
				response.Choices[0] = Choice{
					Text:         responseText,
					Index:        0,
					LogProbs:     nil,
					FinishReason: "length",
				}
				response.Usage = Usage{
					PromptTokens:     int(promptEvalCount),
					CompletionTokens: int(evalCount),
					TotalTokens:      int(promptEvalCount + evalCount),
				}
				jsonData, _ := json.Marshal(response)
				c.Writer.Write([]byte("data: " + string(jsonData) + "\n\n"))
				c.Writer.Flush()
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

	latency := time.Since(startTime).Milliseconds()

	// 更新指标
	h.MetricsCollector.RecordRequest(&types.Request{
		Model:     req.Model,
		Server:    "ollama",
		TokensIn:  int(promptEvalCount),
		TokensOut: int(evalCount),
		LatencyMs: float64(latency),
		Status:    0,
		Timestamp: time.Now(),
	})

	// 保存到存储
	storageReq := &types.Request{
		ID:        response.ID,
		UserID:    "system",
		Model:     req.Model,
		Prompt:    req.Prompt,
		Response:  fullResponse,
		TokensIn:  int(promptEvalCount),
		TokensOut: int(evalCount),
		Server:    "ollama",
		LatencyMs: float64(latency),
		Timestamp: time.Now(),
		Source:    "api",
	}

	if err := h.Storage.SaveRequest(storageReq); err != nil {
		log.Printf("Failed to save generate request: %v", err)
	}

	// 如果不是流式响应，发送完整响应
	if !req.Stream {
		response.Choices[0] = Choice{
			Text:         fullResponse,
			Index:        0,
			LogProbs:     nil,
			FinishReason: "length",
		}
		response.Usage = Usage{
			PromptTokens:     int(promptEvalCount),
			CompletionTokens: int(evalCount),
			TotalTokens:      int(promptEvalCount + evalCount),
		}
		c.JSON(http.StatusOK, response)
	}
}
