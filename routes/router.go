package routes

import (
	"log"

	"github.com/gin-gonic/gin"

	"llm-fw/handlers"
	"llm-fw/types"
)

// SetupRouter 设置路由器
func SetupRouter(ollamaURL string, storage types.Storage, metricsCollector types.MetricsCollector) (*gin.Engine, error) {
	log.Printf("Setting up router with Ollama URL: %s", ollamaURL)
	router := gin.Default()

	// 创建历史记录管理器
	historyManager := storage.NewHistoryManager(100) // 保存最近100条记录
	historyHandler := handlers.NewHistoryHandler(historyManager)

	// 创建模型处理器
	log.Printf("Initializing model handler...")
	modelHandler := handlers.NewModelHandler(ollamaURL, storage, metricsCollector, storage)
	log.Printf("Model handler initialized successfully")

	// 创建生成处理器
	generateHandler := handlers.NewGenerateHandler(ollamaURL, storage, metricsCollector)

	// 创建聊天处理器
	chatHandler := handlers.NewChatHandler(storage, ollamaURL, metricsCollector)

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
		api.GET("/history/search", historyHandler.SearchHistory)
		api.GET("/tags", gin.WrapF(modelHandler.GetTags)) // 使用本地缓存的模型列表提供 Ollama 兼容的接口
		api.DELETE("/models/:name/stats", modelHandler.DeleteModelStats)
	}

	// 静态文件
	router.Static("/static", "templates/static")                 // 静态资源（CSS、JS等）
	router.StaticFile("/", "templates/index.html")               // 主页
	router.StaticFile("/history.html", "templates/history.html") // 历史记录页面

	log.Printf("Router setup completed")
	return router, nil
}
