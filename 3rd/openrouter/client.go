package openrouter

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"llm-fw/common"
)

// Config 定义了 OpenRouter 的配置
type Config struct {
	APIKey  string            `yaml:"api_key"`
	BaseURL string            `yaml:"base_url"`
	Models  map[string]string `yaml:"models"` // 别名到实际模型的映射
}

// Client 是 OpenRouter API 的客户端
type Client struct {
	config Config
	client *http.Client
}

// NewClient 创建一个新的 OpenRouter 客户端
func NewClient(config Config) *Client {
	return &Client{
		config: config,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ChatRequest 定义了聊天请求的结构
type ChatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

// Message 定义了消息的结构
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChatResponse 定义了聊天响应的结构
type ChatResponse struct {
	ID      string   `json:"id"`
	Choices []Choice `json:"choices"`
	Usage   Usage    `json:"usage"`
}

// Choice 定义了选择的结构
type Choice struct {
	Index        int     `json:"index"`
	Delta        Delta   `json:"delta"`
	FinishReason *string `json:"finish_reason"`
}

// Delta 定义了增量更新的结构
type Delta struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Usage 定义了使用统计的结构
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// StreamResponse 定义了流式响应的结构
type StreamResponse struct {
	ID      string   `json:"id"`
	Choices []Choice `json:"choices"`
}

// Chat 发送聊天请求并返回响应
func (c *Client) Chat(modelAlias string, messages []Message) (*ChatResponse, error) {
	// 获取实际的模型名称
	model, ok := c.config.Models[modelAlias]
	if !ok {
		return nil, fmt.Errorf("model alias %s not found in configuration", modelAlias)
	}

	reqBody := ChatRequest{
		Model:    model,
		Messages: messages,
		Stream:   false,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/chat/completions", c.config.BaseURL), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("HTTP-Referer", "https://github.com/yourusername/llm-fw")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(body, &chatResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return &chatResp, nil
}

// ChatStream 发送流式聊天请求并返回响应通道
func (c *Client) ChatStream(modelAlias string, messages []Message) (<-chan string, <-chan error, error) {
	// 获取实际的模型名称
	model, ok := c.config.Models[modelAlias]
	if !ok {
		return nil, nil, fmt.Errorf("model alias %s not found in configuration", modelAlias)
	}

	reqBody := ChatRequest{
		Model:    model,
		Messages: messages,
		Stream:   true,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to marshal request: %v", err)
	}

	fmt.Printf("Request body: %s\n", string(jsonData))

	req, err := http.NewRequest("POST", fmt.Sprintf("%s/v1/chat/completions", c.config.BaseURL), bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.config.APIKey))
	req.Header.Set("HTTP-Referer", "https://github.com/yourusername/llm-fw")
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to send request: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	responseChan := make(chan string, 100)
	errorChan := make(chan error, 1)

	go func() {
		defer resp.Body.Close()
		defer close(responseChan)
		defer close(errorChan)

		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			line := scanner.Text()
			fmt.Printf("Raw response line: %s\n", line) // 打印原始响应行
			if line == "" {
				continue
			}
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				continue
			}

			var streamResp StreamResponse
			if err := json.Unmarshal([]byte(data), &streamResp); err != nil {
				errorChan <- fmt.Errorf("failed to unmarshal stream response: %v, data: %s", err, data)
				return
			}

			if len(streamResp.Choices) > 0 && streamResp.Choices[0].Delta.Content != "" {
				content := streamResp.Choices[0].Delta.Content
				fmt.Printf("Parsed content: %s\n", content)
				responseChan <- content
			}
		}

		if err := scanner.Err(); err != nil {
			errorChan <- fmt.Errorf("failed to read stream: %v", err)
		}
	}()

	return responseChan, errorChan, nil
}

// GetAvailableModels 返回所有可用的模型
func (c *Client) GetAvailableModels() []common.ModelInfo {
	var models []common.ModelInfo
	for alias, model := range c.config.Models {
		models = append(models, common.ModelInfo{
			Name:        alias,
			Family:      "openrouter",
			Parameters:  model,
			IsAvailable: true,
		})
	}
	return models
}
