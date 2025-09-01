package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// MonitorAPI 监控API处理器
type MonitorAPI struct {
	monitor *Monitor
}

// NewMonitorAPI 创建监控API处理器
func NewMonitorAPI(monitor *Monitor) *MonitorAPI {
	return &MonitorAPI{
		monitor: monitor,
	}
}

// RegisterRoutes 注册监控API路由
func (api *MonitorAPI) RegisterRoutes(r *gin.RouterGroup) {
	// 系统概览
	r.GET("/overview", api.GetOverview)

	// 系统监控
	r.GET("/system", api.GetSystemStats)
	r.GET("/system/latest", api.GetLatestSystemStats)

	// SQL分析
	r.GET("/sql/slow", api.GetSlowQueries)
	r.GET("/sql/patterns", api.GetQueryPatterns)
	r.GET("/sql/stats", api.GetSQLStats)
	r.GET("/sql/table/:table", api.GetQueriesByTable)
	r.GET("/sql/operation/:operation", api.GetQueriesByOperation)

	// 链路追踪
	r.GET("/traces", api.GetTraces)
	r.GET("/traces/:traceID", api.GetTraceDetail)

	// 指标数据
	r.GET("/metrics", api.GetMetrics)
	r.GET("/metrics/prometheus", api.GetPrometheusMetrics)
}

// GetOverview 获取系统概览
func (api *MonitorAPI) GetOverview(c *gin.Context) {
	summary := api.monitor.GetSystemSummary()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    summary,
	})
}

// GetSystemStats 获取系统统计
func (api *MonitorAPI) GetSystemStats(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "100")
	limit, _ := strconv.Atoi(limitStr)

	stats := api.monitor.GetSystemStats(limit)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetLatestSystemStats 获取最新系统统计
func (api *MonitorAPI) GetLatestSystemStats(c *gin.Context) {
	stats := api.monitor.GetLatestSystemStats()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetSlowQueries 获取慢查询列表
func (api *MonitorAPI) GetSlowQueries(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)

	queries := api.monitor.GetSlowQueries(limit)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    queries,
	})
}

// GetQueryPatterns 获取查询模式
func (api *MonitorAPI) GetQueryPatterns(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)

	patterns := api.monitor.GetQueryPatterns(limit)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    patterns,
	})
}

// GetSQLStats 获取SQL统计信息
func (api *MonitorAPI) GetSQLStats(c *gin.Context) {
	if api.monitor.GetSQLAnalyzer() == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    nil,
		})
		return
	}

	stats := api.monitor.GetSQLAnalyzer().GetQueryStats()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetQueriesByTable 按表获取查询
func (api *MonitorAPI) GetQueriesByTable(c *gin.Context) {
	table := c.Param("table")
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)

	queries := api.monitor.GetQueriesByTable(table, limit)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    queries,
	})
}

// GetQueriesByOperation 按操作类型获取查询
func (api *MonitorAPI) GetQueriesByOperation(c *gin.Context) {
	operation := c.Param("operation")
	limitStr := c.DefaultQuery("limit", "50")
	limit, _ := strconv.Atoi(limitStr)

	queries := api.monitor.GetQueriesByOperation(operation, limit)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    queries,
	})
}

// GetTraces 获取追踪列表
func (api *MonitorAPI) GetTraces(c *gin.Context) {
	if api.monitor.GetTracer() == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []interface{}{},
		})
		return
	}

	spans := api.monitor.GetTracer().GetSpans()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    spans,
	})
}

// GetTraceDetail 获取追踪详情
func (api *MonitorAPI) GetTraceDetail(c *gin.Context) {
	traceID := c.Param("traceID")

	if api.monitor.GetTracer() == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    nil,
		})
		return
	}

	spans := api.monitor.GetTraceSpans(traceID)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    spans,
	})
}

// GetMetrics 获取指标数据
func (api *MonitorAPI) GetMetrics(c *gin.Context) {
	if api.monitor.GetMetrics() == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    nil,
		})
		return
	}

	// 这里可以返回自定义的指标数据
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"timestamp": time.Now(),
			"message":   "Metrics collection is enabled",
		},
	})
}

// GetPrometheusMetrics 获取Prometheus格式的指标
func (api *MonitorAPI) GetPrometheusMetrics(c *gin.Context) {
	// 这里应该返回Prometheus格式的指标数据
	// 由于Prometheus指标是自动注册的，我们只需要返回一个说明
	c.Header("Content-Type", "text/plain")
	c.String(http.StatusOK, "# Prometheus metrics are automatically exposed at /metrics endpoint\n# This endpoint is for compatibility only")
}
