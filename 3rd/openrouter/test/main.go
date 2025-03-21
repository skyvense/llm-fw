package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"

	"llm-fw/3rd/openrouter"
	"llm-fw/common"
)

func selectModel(models []common.ModelInfo) string {
	fmt.Println("\nAvailable models:")
	for i, model := range models {
		fmt.Printf("%d. %s (%s)\n", i+1, model.Name, model.Parameters)
	}

	var choice int
	for {
		fmt.Print("\nSelect a model (enter number): ")
		_, err := fmt.Scanf("%d", &choice)
		if err != nil {
			fmt.Println("Invalid input. Please enter a number.")
			continue
		}
		if choice < 1 || choice > len(models) {
			fmt.Printf("Please enter a number between 1 and %d.\n", len(models))
			continue
		}
		break
	}

	return models[choice-1].Name
}

func main() {
	// 解析命令行参数
	message := flag.String("message", "Hello", "Message to send to the model")
	flag.Parse()

	// 获取当前工作目录
	workDir, err := os.Getwd()
	if err != nil {
		log.Fatalf("Failed to get working directory: %v", err)
	}

	// 读取配置文件
	configPath := filepath.Join(workDir, "config.yaml")
	configData, err := os.ReadFile(configPath)
	if err != nil {
		log.Fatalf("Failed to read config file at %s: %v", configPath, err)
	}

	var config struct {
		OpenRouter openrouter.Config `yaml:"openrouter"`
	}
	if err := yaml.Unmarshal(configData, &config); err != nil {
		log.Fatalf("Failed to parse config file: %v", err)
	}

	// 检查配置
	if config.OpenRouter.APIKey == "" {
		log.Fatal("OpenRouter API key is not set in config.yaml")
	}
	if config.OpenRouter.BaseURL == "" {
		log.Fatal("OpenRouter base URL is not set in config.yaml")
	}
	if len(config.OpenRouter.Models) == 0 {
		log.Fatal("No models configured in config.yaml")
	}

	fmt.Printf("Using OpenRouter API at: %s\n", config.OpenRouter.BaseURL)
	fmt.Printf("Number of configured models: %d\n", len(config.OpenRouter.Models))

	// 创建客户端
	client := openrouter.NewClient(config.OpenRouter)

	// 获取可用模型并选择
	models := client.GetAvailableModels()
	selectedModel := selectModel(models)

	// 准备消息
	messages := []openrouter.Message{
		{
			Role:    "user",
			Content: *message,
		},
	}

	// 发送流式请求
	fmt.Printf("\nSending request to model: %s\n", selectedModel)
	responseChan, errorChan, err := client.ChatStream(selectedModel, messages)
	if err != nil {
		log.Fatalf("Failed to start chat stream: %v", err)
	}

	// 处理响应
	fmt.Printf("\nResponse from %s:\n", selectedModel)
	fmt.Println(strings.Repeat("-", 50))

	var fullResponse strings.Builder
	for {
		select {
		case chunk, ok := <-responseChan:
			if !ok {
				fmt.Println("\n" + strings.Repeat("-", 50))
				fmt.Printf("Full response:\n%s\n", fullResponse.String())
				return
			}
			fmt.Print(chunk)
			fullResponse.WriteString(chunk)
		case err := <-errorChan:
			if err != nil {
				log.Printf("Error in chat stream: %v", err)
				fmt.Printf("Partial response:\n%s\n", fullResponse.String())
				return
			}
		}
	}
}
