package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"llm-fw/api"
	"llm-fw/interfaces"
)

// HistoryHandler 处理历史记录相关的请求
type HistoryHandler struct {
	historyManager interfaces.HistoryManager
}

// NewHistoryHandler 创建一个新的历史记录处理器
func NewHistoryHandler(historyManager interfaces.HistoryManager) *HistoryHandler {
	return &HistoryHandler{
		historyManager: historyManager,
	}
}

// GetHistory 获取历史记录
func (h *HistoryHandler) GetHistory(c *gin.Context) {
	if c.Request.Method != http.MethodGet {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method not allowed"})
		return
	}

	entries, err := h.historyManager.GetHistory()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"entries": entries})
}

// AddEntry 添加一条历史记录
func (h *HistoryHandler) AddEntry(entry api.HistoryEntry) error {
	return h.historyManager.AddEntry(entry)
}

// ClearHistory 清空历史记录
func (h *HistoryHandler) ClearHistory(c *gin.Context) {
	if c.Request.Method != http.MethodDelete {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method not allowed"})
		return
	}

	if err := h.historyManager.ClearHistory(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear history"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "History cleared successfully"})
}
