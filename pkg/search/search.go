package search

import (
	"HibiscusIM/pkg/config"
	"HibiscusIM/pkg/response"
	"log"

	"github.com/gin-gonic/gin"
)

// SearchHandlers 封装搜索相关的API处理
type SearchHandlers struct {
	engine Engine
}

// NewSearchHandlers 创建一个新的SearchHandlers实例
func NewSearchHandlers(engine Engine) *SearchHandlers {
	return &SearchHandlers{
		engine: engine,
	}
}

// RegisterSearchRoutes 注册与搜索相关的路由
func (h *SearchHandlers) RegisterSearchRoutes(r *gin.RouterGroup) {
	if !config.GlobalConfig.SearchEnabled {
		log.Println("Search feature is disabled")
		return
	}

	// Search API 路由
	searchGroup := r.Group("/search")
	{
		// 搜索接口
		searchGroup.POST("/", h.handleSearch)
		// 索引文档接口
		searchGroup.POST("/index", h.handleIndex)
		// 删除文档接口
		searchGroup.POST("/delete", h.handleDelete)
		// 自动补全接口
		searchGroup.POST("/auto-complete", h.handleAutoComplete)
		// 搜索建议接口
		searchGroup.POST("/suggest", h.handleSuggest)
	}
}

// handleSearch 处理搜索请求
func (h *SearchHandlers) handleSearch(c *gin.Context) {
	var req SearchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "Invalid search request", gin.H{"error": err.Error()})
		return
	}

	// 执行搜索
	result, err := h.engine.Search(c, req)
	if err != nil {
		response.Fail(c, "Internal Server Error", gin.H{"error": err.Error()})
		return
	}

	response.Success(c, "Get Search Result", result)
}

// handleIndex 处理文档索引请求
func (h *SearchHandlers) handleIndex(c *gin.Context) {
	var doc Doc
	if err := c.ShouldBindJSON(&doc); err != nil {
		response.Fail(c, "Invalid document", gin.H{"error": err.Error()})
		return
	}

	// 索引文档
	err := h.engine.Index(c, doc)
	if err != nil {
		response.Fail(c, "Internal Server Error", gin.H{"error": err.Error()})
		return
	}
	response.Success(c, "Document indexed successfully", gin.H{"doc": doc})
}

// handleDelete 处理文档删除请求
func (h *SearchHandlers) handleDelete(c *gin.Context) {
	var req struct {
		ID string `json:"id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "Invalid document", gin.H{"error": err.Error()})
		return
	}

	// 删除文档
	err := h.engine.Delete(c, req.ID)
	if err != nil {
		response.Fail(c, "Internal Server Error", gin.H{"error": err.Error()})
		return
	}
	response.Success(c, "Document deleted successfully", nil)
}

// handleAutoComplete 处理自动补全请求
func (h *SearchHandlers) handleAutoComplete(c *gin.Context) {
	var req struct {
		Keyword string `json:"keyword"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "Invalid keyword", gin.H{"error": err.Error()})
		return
	}

	// 获取自动补全建议
	suggestions, err := h.engine.GetAutoCompleteSuggestions(c, req.Keyword)
	if err != nil {
		response.Fail(c, "Internal Server Error", gin.H{"error": err.Error()})
		return
	}
	response.Success(c, "Get Suggestion successfully", suggestions)
}

// handleSuggest 处理搜索建议请求
func (h *SearchHandlers) handleSuggest(c *gin.Context) {
	var req struct {
		Keyword string `json:"keyword"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "Invalid keyword", gin.H{"error": err.Error()})
		return
	}

	// 获取基于关键词的搜索建议
	suggestions, err := h.engine.GetSearchSuggestions(c, req.Keyword)
	if err != nil {
		response.Fail(c, "Internal Server Error", gin.H{"error": err.Error()})
		return
	}

	response.Success(c, "Get Suggestion successfully", suggestions)
}
