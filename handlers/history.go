package handlers

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"llm-fw/common"
	"llm-fw/types"
)

// HistoryHandler 处理历史记录相关的请求
type HistoryHandler struct {
	historyManager types.HistoryManager
}

// NewHistoryHandler 创建一个新的历史记录处理器
func NewHistoryHandler(historyManager types.HistoryManager) *HistoryHandler {
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

	// 设置 CORS 头
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type")
	c.Header("Vary", "Origin")

	// 处理 OPTIONS 请求
	if c.Request.Method == http.MethodOptions {
		c.Status(http.StatusOK)
		return
	}

	// 获取限制参数
	limit := 5 // 默认显示5条
	if limitStr := c.Query("limit"); limitStr != "" {
		if parsedLimit, err := strconv.Atoi(limitStr); err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	requests := h.historyManager.Get()

	// 确保按时间戳降序排序
	sort.Slice(requests, func(i, j int) bool {
		return requests[i].Timestamp.After(requests[j].Timestamp)
	})

	// 限制返回数量
	if len(requests) > limit {
		requests = requests[:limit]
	}

	c.JSON(http.StatusOK, gin.H{"requests": requests})
}

// SearchHistory 搜索历史记录
func (h *HistoryHandler) SearchHistory(c *gin.Context) {
	// 设置 CORS 头
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, OPTIONS")
	c.Header("Access-Control-Allow-Headers", "Content-Type")
	c.Header("Vary", "Origin")

	// 处理 OPTIONS 请求
	if c.Request.Method == http.MethodOptions {
		c.Status(http.StatusOK)
		return
	}

	// 获取查询参数
	model := c.Query("model")
	keyword := c.Query("keyword")
	startDateStr := c.Query("start_date")
	endDateStr := c.Query("end_date")
	pageStr := c.Query("page")
	pageSizeStr := c.Query("page_size")

	// 解析日期
	var startDate, endDate time.Time
	var err error
	if startDateStr != "" {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format"})
			return
		}
	}
	if endDateStr != "" {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format"})
			return
		}
		// 将结束日期设置为当天的最后一秒
		endDate = endDate.Add(24*time.Hour - time.Second)
	}

	// 解析分页参数
	page := 1
	if pageStr != "" {
		if parsedPage, err := strconv.Atoi(pageStr); err == nil && parsedPage > 0 {
			page = parsedPage
		}
	}

	pageSize := 20 // 默认每页20条
	if pageSizeStr != "" {
		if parsedPageSize, err := strconv.Atoi(pageSizeStr); err == nil && parsedPageSize > 0 {
			pageSize = parsedPageSize
		}
	}

	// 构建搜索参数
	params := types.SearchParams{
		Model:     model,
		StartDate: startDate,
		EndDate:   endDate,
		Keyword:   keyword,
		Page:      page,
		PageSize:  pageSize,
	}

	// 执行搜索
	result, err := h.historyManager.Search(params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"requests": result.Requests,
		"total":    result.Total,
	})
}

// AddEntry 添加一条历史记录
func (h *HistoryHandler) AddEntry(entry *common.Request) error {
	h.historyManager.Add(entry)
	return nil
}

// ClearHistory 清空历史记录
func (h *HistoryHandler) ClearHistory(c *gin.Context) {
	if c.Request.Method != http.MethodDelete {
		c.JSON(http.StatusMethodNotAllowed, gin.H{"error": "Method not allowed"})
		return
	}

	h.historyManager.Clear()

	c.JSON(http.StatusOK, gin.H{"message": "History cleared successfully"})
}
