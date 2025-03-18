package routes

import (
	"log"
	"net/http/httputil"
	"net/url"

	"github.com/gin-gonic/gin"

	"llm-fw/handlers"
	"llm-fw/interfaces"
)

// SetupRouter 设置路由器
func SetupRouter(ollamaURL string, storage interfaces.Storage, metricsCollector interfaces.MetricsCollector) (*gin.Engine, error) {
	log.Printf("Setting up router with Ollama URL: %s", ollamaURL)
	router := gin.Default()

	// 创建历史记录管理器
	historyManager := storage.NewHistoryManager(100) // 保存最近100条记录
	historyHandler := handlers.NewHistoryHandler(historyManager)

	// 创建模型处理器
	log.Printf("Initializing model handler...")
	modelHandler, err := handlers.NewModelHandler(ollamaURL, metricsCollector)
	if err != nil {
		log.Printf("Failed to initialize model handler: %v", err)
		return nil, err
	}
	log.Printf("Model handler initialized successfully")

	// 创建生成处理器
	generateHandler := handlers.NewGenerateHandler(ollamaURL, storage, metricsCollector)

	// 创建聊天处理器
	chatHandler := handlers.NewChatHandler(ollamaURL, storage, metricsCollector)

	// 设置 Ollama 代理
	ollamaTarget, err := url.Parse(ollamaURL)
	if err != nil {
		return nil, err
	}
	ollamaProxy := httputil.NewSingleHostReverseProxy(ollamaTarget)

	// API 路由组
	api := router.Group("/api")
	{
		// 生成相关路由
		api.POST("/generate", generateHandler.Generate)

		// 聊天相关路由
		api.POST("/chat", chatHandler.Chat)

		// 模型相关路由
		api.GET("/models", gin.WrapF(modelHandler.ListModels))
		api.GET("/history", historyHandler.GetHistory)

		// Ollama API 代理路由
		api.Any("/tags", func(c *gin.Context) {
			ollamaProxy.ServeHTTP(c.Writer, c.Request)
		})
	}

	// 静态文件
	router.Static("/static", "templates/static")   // 静态资源（CSS、JS等）
	router.StaticFile("/", "templates/index.html") // 主页

	log.Printf("Router setup completed")
	return router, nil
}
