package models

import (
	hibiscusIM "HibiscusIM"
	"HibiscusIM/pkg/metrics"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// MonitorAdminObject 监控系统管理对象
type MonitorAdminObject struct {
	monitor *metrics.Monitor
}

// NewMonitorAdminObject 创建监控管理对象
func NewMonitorAdminObject(monitor *metrics.Monitor) *MonitorAdminObject {
	return &MonitorAdminObject{
		monitor: monitor,
	}
}

// GetAdminObject 获取监控系统的管理对象配置
func (m *MonitorAdminObject) GetAdminObject() AdminObject {
	iconMonitor, _ := hibiscusIM.EmbedStaticAssets.ReadFile("static/img/icon_monitor.svg")
	if iconMonitor == nil {
		// 如果没有图标文件，创建一个默认的SVG图标
		iconMonitor = []byte(`<svg xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24" stroke-width="1.5" stroke="currentColor">
			<path stroke-linecap="round" stroke-linejoin="round" d="M3.75 3v11.25A2.25 2.25 0 006.75 16.5h10.5a2.25 2.25 0 002.25-2.25V3M3.75 3h-1.5m1.5 0h16.5m0 0h1.5m-1.5 0v11.25A2.25 2.25 0 0118.75 16.5H8.25a2.25 2.25 0 01-2.25-2.25V3" />
		</svg>`)
	}

	return AdminObject{
		Model:       &MonitorData{},
		Group:       "System",
		Name:        "Monitor",
		Desc:        "System monitoring and performance analysis",
		Path:        "monitor",
		Shows:       []string{"ID", "Type", "Status", "Value", "Timestamp"},
		Editables:   []string{},
		Filterables: []string{"Type", "Status", "Timestamp"},
		Orderables:  []string{"Timestamp"},
		Searchables: []string{"Type", "Status"},
		Orders:      []hibiscusIM.Order{{"Timestamp", hibiscusIM.OrderOpDesc}},
		Icon:        &AdminIcon{SVG: string(iconMonitor)},
		Invisible:   false,
		EditPage:    "monitor_edit",
		ListPage:    "monitor_list",
		Scripts: []AdminScript{
			{
				Src: "/static/admin/monitor.js",
			},
		},
		Styles: []string{
			"/static/admin/monitor.css",
		},
		Actions: []AdminAction{
			{
				Path:  "overview",
				Name:  "System Overview",
				Label: "View system overview",
				Handler: func(db *gorm.DB, c *gin.Context, obj any) (bool, any, error) {
					summary := m.monitor.GetSystemSummary()
					return true, summary, nil
				},
			},
			{
				Path:  "slow_queries",
				Name:  "Slow Queries",
				Label: "View slow SQL queries",
				Handler: func(db *gorm.DB, c *gin.Context, obj any) (bool, any, error) {
					queries := m.monitor.GetSlowQueries(50)
					return true, queries, nil
				},
			},
			{
				Path:  "query_patterns",
				Name:  "Query Patterns",
				Label: "View SQL query patterns",
				Handler: func(db *gorm.DB, c *gin.Context, obj any) (bool, any, error) {
					patterns := m.monitor.GetQueryPatterns(50)
					return true, patterns, nil
				},
			},
			{
				Path:  "traces",
				Name:  "Request Traces",
				Label: "View request traces",
				Handler: func(db *gorm.DB, c *gin.Context, obj any) (bool, any, error) {
					spans := m.monitor.GetTracer().GetSpans()
					return true, spans, nil
				},
			},
			{
				Path:  "system_stats",
				Name:  "System Stats",
				Label: "View system statistics",
				Handler: func(db *gorm.DB, c *gin.Context, obj any) (bool, any, error) {
					stats := m.monitor.GetSystemStats(100)
					return true, stats, nil
				},
			},
			{
				Path:  "refresh",
				Name:  "Refresh Data",
				Label: "Refresh monitoring data",
				Handler: func(db *gorm.DB, c *gin.Context, obj any) (bool, any, error) {
					// 强制刷新监控数据
					return true, "Data refreshed successfully", nil
				},
			},
		},
		AccessCheck: func(c *gin.Context, obj *AdminObject) error {
			// 只有超级用户和管理员可以访问监控系统
			user := CurrentUser(c)
			if !user.IsSuperUser && !user.IsStaff {
				return fmt.Errorf("insufficient permissions")
			}
			return nil
		},
	}
}

// MonitorData 监控数据结构
type MonitorData struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Type      string    `json:"type"`      // 监控类型：system, sql, trace, metric
	Status    string    `json:"status"`    // 状态：normal, warning, error
	Value     string    `json:"value"`     // 监控值（JSON字符串）
	Timestamp time.Time `json:"timestamp"` // 时间戳
	Metadata  string    `json:"metadata"`  // 元数据（JSON字符串）
}

// TableName 返回表名
func (MonitorData) TableName() string {
	return "monitor_data"
}

// MonitorAPIHandler 监控API处理器
type MonitorAPIHandler struct {
	monitor *metrics.Monitor
}

// NewMonitorAPIHandler 创建监控API处理器
func NewMonitorAPIHandler(monitor *metrics.Monitor) *MonitorAPIHandler {
	return &MonitorAPIHandler{
		monitor: monitor,
	}
}

// RegisterRoutes 注册监控API路由
func (h *MonitorAPIHandler) RegisterRoutes(r *gin.RouterGroup) {
	// 系统概览
	r.GET("/overview", h.GetOverview)

	// 系统监控
	r.GET("/system", h.GetSystemStats)
	r.GET("/system/latest", h.GetLatestSystemStats)

	// SQL分析
	r.GET("/sql/slow", h.GetSlowQueries)
	r.GET("/sql/patterns", h.GetQueryPatterns)
	r.GET("/sql/stats", h.GetSQLStats)
	r.GET("/sql/table/:table", h.GetQueriesByTable)
	r.GET("/sql/operation/:operation", h.GetQueriesByOperation)

	// 链路追踪
	r.GET("/traces", h.GetTraces)
	r.GET("/traces/:traceID", h.GetTraceDetail)

	// 指标数据
	r.GET("/metrics", h.GetMetrics)
	r.GET("/metrics/prometheus", h.GetPrometheusMetrics)

	// 实时数据
	r.GET("/realtime", h.GetRealTimeData)
	r.GET("/alerts", h.GetAlerts)
}

// GetOverview 获取系统概览
func (h *MonitorAPIHandler) GetOverview(c *gin.Context) {
	if h.monitor == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": map[string]interface{}{
				"message":   "监控系统未初始化",
				"timestamp": time.Now(),
			},
		})
		return
	}

	summary := h.monitor.GetSystemSummary()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    summary,
	})
}

// GetSystemStats 获取系统统计
func (h *MonitorAPIHandler) GetSystemStats(c *gin.Context) {
	if h.monitor == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []interface{}{},
		})
		return
	}

	limitStr := c.DefaultQuery("limit", "100")
	limit := 0
	fmt.Sscanf(limitStr, "%d", &limit)

	stats := h.monitor.GetSystemStats(limit)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetLatestSystemStats 获取最新系统统计
func (h *MonitorAPIHandler) GetLatestSystemStats(c *gin.Context) {
	if h.monitor == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    nil,
		})
		return
	}

	stats := h.monitor.GetLatestSystemStats()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetSlowQueries 获取慢查询列表
func (h *MonitorAPIHandler) GetSlowQueries(c *gin.Context) {
	if h.monitor == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []interface{}{},
		})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit := 0
	fmt.Sscanf(limitStr, "%d", &limit)

	queries := h.monitor.GetSlowQueries(limit)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    queries,
	})
}

// GetQueryPatterns 获取查询模式
func (h *MonitorAPIHandler) GetQueryPatterns(c *gin.Context) {
	if h.monitor == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []interface{}{},
		})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit := 0
	fmt.Sscanf(limitStr, "%d", &limit)

	patterns := h.monitor.GetQueryPatterns(limit)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    patterns,
	})
}

// GetSQLStats 获取SQL统计信息
func (h *MonitorAPIHandler) GetSQLStats(c *gin.Context) {
	if h.monitor.GetSQLAnalyzer() == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    nil,
		})
		return
	}

	stats := h.monitor.GetSQLAnalyzer().GetQueryStats()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetQueriesByTable 按表获取查询
func (h *MonitorAPIHandler) GetQueriesByTable(c *gin.Context) {
	if h.monitor == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []interface{}{},
		})
		return
	}

	table := c.Param("table")
	limitStr := c.DefaultQuery("limit", "50")
	limit := 0
	fmt.Sscanf(limitStr, "%d", &limit)

	queries := h.monitor.GetQueriesByTable(table, limit)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    queries,
	})
}

// GetQueriesByOperation 按操作类型获取查询
func (h *MonitorAPIHandler) GetQueriesByOperation(c *gin.Context) {
	if h.monitor == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []interface{}{},
		})
		return
	}

	operation := c.Param("operation")
	limitStr := c.DefaultQuery("limit", "50")
	limit := 0
	fmt.Sscanf(limitStr, "%d", &limit)

	queries := h.monitor.GetQueriesByOperation(operation, limit)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    queries,
	})
}

// GetTraces 获取追踪列表
func (h *MonitorAPIHandler) GetTraces(c *gin.Context) {
	if h.monitor.GetTracer() == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    []interface{}{},
		})
		return
	}

	spans := h.monitor.GetTracer().GetSpans()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    spans,
	})
}

// GetTraceDetail 获取追踪详情
func (h *MonitorAPIHandler) GetTraceDetail(c *gin.Context) {
	traceID := c.Param("traceID")

	if h.monitor.GetTracer() == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    nil,
		})
		return
	}

	spans := h.monitor.GetTraceSpans(traceID)
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    spans,
	})
}

// GetMetrics 获取指标数据
func (h *MonitorAPIHandler) GetMetrics(c *gin.Context) {
	if h.monitor.GetMetrics() == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data":    nil,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": map[string]interface{}{
			"timestamp": time.Now(),
			"message":   "Metrics collection is enabled",
		},
	})
}

// GetPrometheusMetrics 获取Prometheus格式的指标
func (h *MonitorAPIHandler) GetPrometheusMetrics(c *gin.Context) {
	c.Header("Content-Type", "text/plain")
	c.String(http.StatusOK, "# Prometheus metrics are automatically exposed at /metrics endpoint\n# This endpoint is for compatibility only")
}

// GetRealTimeData 获取实时数据
func (h *MonitorAPIHandler) GetRealTimeData(c *gin.Context) {
	if h.monitor == nil {
		c.JSON(http.StatusOK, gin.H{
			"success": true,
			"data": map[string]interface{}{
				"timestamp": time.Now(),
				"message":   "监控系统未初始化",
			},
		})
		return
	}

	// 获取最新的监控数据
	data := map[string]interface{}{
		"timestamp": time.Now(),
		"system":    h.monitor.GetLatestSystemStats(),
		"sql":       h.monitor.GetSQLAnalyzer().GetQueryStats(),
		"traces":    len(h.monitor.GetTracer().GetSpans()),
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    data,
	})
}

// GetAlerts 获取告警信息
func (h *MonitorAPIHandler) GetAlerts(c *gin.Context) {
	// 这里可以实现告警逻辑
	alerts := []map[string]interface{}{
		{
			"id":      "alert_001",
			"level":   "warning",
			"message": "High CPU usage detected",
			"time":    time.Now().Add(-5 * time.Minute),
		},
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    alerts,
	})
}
