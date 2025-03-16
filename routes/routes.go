package routes

import (
	"llm-fw/handlers"
	"llm-fw/proxy"
	"time"

	"github.com/gin-gonic/gin"
)

// SetupRouter configures all routes
func SetupRouter(
	targetURL string,
	storage handlers.Storage,
	metricsCollector handlers.MetricsCollector,
) (*gin.Engine, error) {
	r := gin.Default()

	// Load templates
	r.LoadHTMLGlob("templates/*")

	// Web interface routes
	r.GET("/", func(c *gin.Context) {
		c.HTML(200, "dashboard.html", nil)
	})

	// Create reverse proxy
	reverseProxy, err := proxy.NewReverseProxy(targetURL)
	if err != nil {
		return nil, err
	}

	// Create handlers
	modelHandler := handlers.NewModelHandler(targetURL, metricsCollector)
	generateHandler := handlers.NewGenerateHandler(targetURL, storage, metricsCollector)
	chatHandler := handlers.NewChatHandler(targetURL, storage, metricsCollector)

	// API routes
	api := r.Group("/api")
	{
		// Direct proxy to Ollama for model operations
		api.POST("/generate", generateHandler.Generate)
		api.POST("/chat", chatHandler.Chat)
		api.GET("/models", modelHandler.ListModels)
		api.GET("/models/:model/stats", modelHandler.GetModelStats)
	}

	// Proxy all other Ollama API requests
	r.NoRoute(func(c *gin.Context) {
		startTime := time.Now()
		reverseProxy.ServeHTTP(c.Writer, c.Request)
		latency := time.Since(startTime).Milliseconds()

		metricsCollector.RecordRequest(
			"ollama",
			"proxy",
			int64(len(c.Request.URL.Path)),
			int64(len(c.Writer.Header().Get("Content-Type"))),
			latency,
			c.Writer.Status() < 400,
		)
	})

	return r, nil
}
