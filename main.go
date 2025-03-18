package main

import (
	"fmt"
	"log"

	"llm-fw/config"
	"llm-fw/metrics"
	"llm-fw/routes"
	"llm-fw/storage"
)

func main() {
	// 初始化配置
	cfg := config.NewConfig()

	// 尝试加载配置文件
	if loadedCfg, err := config.LoadConfig("config.yaml"); err == nil {
		cfg = loadedCfg
		log.Println("已加载配置文件 config.yaml")
	} else {
		log.Printf("未找到配置文件，使用默认配置: %v", err)
	}

	// 初始化存储
	storage, err := storage.NewStorage(cfg)
	if err != nil {
		log.Fatalf("初始化存储失败: %v", err)
	}
	defer storage.Close()

	// 初始化指标收集器
	metricsCollector := metrics.NewMetrics()

	// 设置路由
	router, err := routes.SetupRouter(cfg.Ollama.URL, storage, metricsCollector)
	if err != nil {
		log.Fatalf("设置路由失败: %v", err)
	}

	// 启动服务器
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("服务器启动在 %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("服务器启动失败: %v", err)
	}
}
