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
	"llm-fw/routes"
	"llm-fw/storage"
	"llm-fw/types"
)

func main() {
	// Load configuration
	cfg, err := config.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize storage
	var store types.Storage
	switch cfg.Storage.Type {
	case "file":
		store, err = storage.NewFileStorageImpl(cfg.Storage.Path)
	case "sqlite":
		store, err = storage.NewSQLiteStorage(cfg.Storage.Path)
	default:
		log.Fatalf("Unsupported storage type: %s", cfg.Storage.Type)
	}
	if err != nil {
		log.Fatalf("Failed to initialize storage: %v", err)
	}
	defer store.Close()

	// Initialize metrics collector
	metricsCollector := &types.NoopMetricsCollector{}

	// Create router
	router, err := routes.SetupRouter(cfg.Ollama.URL, store, metricsCollector)
	if err != nil {
		log.Fatalf("Failed to setup router: %v", err)
	}

	// Create server
	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler: router,
	}

	// Start server
	go func() {
		log.Printf("Server starting on %s:%d", cfg.Server.Host, cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	// Graceful shutdown
	log.Println("Shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
}
