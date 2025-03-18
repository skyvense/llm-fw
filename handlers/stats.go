package handlers

import (
	"llm-fw/metrics"
	"llm-fw/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

type StatsHandler struct {
	storage          types.Storage
	metricsCollector *metrics.Metrics
}

func NewStatsHandler(storage types.Storage, metricsCollector *metrics.Metrics) *StatsHandler {
	return &StatsHandler{
		storage:          storage,
		metricsCollector: metricsCollector,
	}
}

// GetStats 获取统计信息
func (h *StatsHandler) GetStats(c *gin.Context) {
	// 获取所有模型的统计信息
	modelStats, err := h.storage.GetAllModelStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取模型统计失败",
		})
		return
	}

	// 获取最近的请求记录
	recentRequests, err := h.storage.GetRecentRequests(10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取最近请求失败",
		})
		return
	}

	// 获取指标数据
	metrics := h.metricsCollector.GetMetrics()

	c.JSON(http.StatusOK, gin.H{
		"model_stats":      modelStats,
		"recent_requests":  recentRequests,
		"server_health":    metrics.ServerHealth,
		"total_requests":   metrics.TotalRequests,
		"total_tokens_in":  metrics.TotalTokensIn,
		"total_tokens_out": metrics.TotalTokensOut,
		"failed_requests":  metrics.FailedRequests,
	})
}

// GetModels 获取模型列表
func (h *StatsHandler) GetModels(c *gin.Context) {
	// 获取所有模型的统计信息
	modelStats, err := h.storage.GetAllModelStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取模型列表失败",
		})
		return
	}

	// 构建模型列表
	var models []string
	for model := range modelStats {
		models = append(models, model)
	}

	c.JSON(http.StatusOK, gin.H{
		"models": models,
	})
}
