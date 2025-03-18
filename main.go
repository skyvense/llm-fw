package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"llm-fw/config"
	"llm-fw/metrics"
	"llm-fw/routes"
	"llm-fw/storage"
)

func main() {
	// 加载配置文件
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化文件存储
	fileStorage, err := storage.NewFileStorage(cfg.Storage.Path)
	if err != nil {
		log.Fatalf("Failed to initialize file storage: %v", err)
	}

	// 初始化存储适配器
	storageAdapter := storage.NewStorageAdapter(fileStorage)

	// 初始化指标收集器
	metricsCollector := metrics.NewMetrics()

	// 设置路由
	router, err := routes.SetupRouter(cfg.Ollama.URL, storageAdapter, metricsCollector)
	if err != nil {
		log.Fatalf("Failed to set up router: %v", err)
	}

	// 创建服务器
	server := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	// 优雅关闭
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP server shutdown error: %v", err)
		}
	}()

	fmt.Printf("Server is running on http://%s:%d\n", cfg.Server.Host, cfg.Server.Port)
	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("HTTP server error: %v", err)
	}
}
