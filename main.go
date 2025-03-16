package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"llm-fw/config"
	"llm-fw/metrics"
	"llm-fw/routes"
	"llm-fw/storage"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize storage
	fileStorage, err := storage.NewFileStorage(cfg.Storage.Path)
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}

	// Create storage adapter
	storageAdapter := storage.NewStorageAdapter(fileStorage)

	// Set up metrics collector
	metricsCollector := metrics.NewMetrics()

	// Start server health check
	go func() {
		for {
			isHealthy := checkServerHealth(cfg.Ollama.URL)
			metricsCollector.UpdateServerHealth("ollama", isHealthy)
			time.Sleep(30 * time.Second)
		}
	}()

	// Set up router
	router, err := routes.SetupRouter(cfg.Ollama.URL, storageAdapter, metricsCollector)
	if err != nil {
		log.Fatalf("Failed to set up router: %v", err)
	}

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Server started on %s", addr)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

func checkServerHealth(url string) bool {
	resp, err := http.Get(fmt.Sprintf("%s/api/health", url))
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}
