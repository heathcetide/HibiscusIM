package metrics

import (
	"context"
	"testing"
	"time"
)

// TestMonitorBasic 测试监控器基本功能
func TestMonitorBasic(t *testing.T) {
	// 创建监控器
	monitor := NewMonitor(nil)

	// 启动监控
	monitor.Start()
	defer monitor.Stop()

	// 检查监控状态
	if !monitor.IsEnabled() {
		t.Error("监控器应该启用")
	}

	// 获取配置
	config := monitor.GetConfig()
	if config == nil {
		t.Error("配置不应该为空")
	}

	// 检查默认配置
	if !config.EnableMetrics {
		t.Error("指标收集应该默认启用")
	}
	if !config.EnableTracing {
		t.Error("链路追踪应该默认启用")
	}
	if !config.EnableSQLAnalysis {
		t.Error("SQL分析应该默认启用")
	}
	if !config.EnableSystemMonitor {
		t.Error("系统监控应该默认启用")
	}
}

// TestTracing 测试链路追踪功能
func TestTracing(t *testing.T) {
	monitor := NewMonitor(nil)
	monitor.Start()
	defer monitor.Stop()

	ctx := context.Background()

	// 开始链路追踪
	ctx, span := monitor.StartSpan(ctx, "test_operation",
		WithTags(map[string]string{
			"test": "true",
		}),
	)

	if span == nil {
		t.Error("span不应该为空")
	}

	// 添加事件
	span.AddEvent("test_event", map[string]interface{}{
		"message": "test event",
	})

	// 设置属性
	span.SetAttribute("test_attr", "test_value")

	// 结束追踪
	monitor.EndSpan(span, nil)

	// 获取追踪
	spans := monitor.GetTracer().GetSpans()
	if len(spans) == 0 {
		t.Error("应该有追踪记录")
	}
}

// TestSQLAnalysis 测试SQL分析功能
func TestSQLAnalysis(t *testing.T) {
	monitor := NewMonitor(nil)
	monitor.Start()
	defer monitor.Stop()

	ctx := context.Background()

	// 记录SQL查询
	monitor.RecordSQLQuery(ctx, "SELECT * FROM users WHERE id = ?",
		[]interface{}{123}, "users", "SELECT", 50*time.Millisecond, 1, nil)

	// 记录慢查询
	monitor.RecordSQLQuery(ctx, "SELECT * FROM users WHERE name LIKE ?",
		[]interface{}{"%test%"}, "users", "SELECT", 150*time.Millisecond, 10, nil)

	// 获取慢查询
	slowQueries := monitor.GetSlowQueries(10)
	if len(slowQueries) == 0 {
		t.Error("应该有慢查询记录")
	}

	// 获取查询模式
	patterns := monitor.GetQueryPatterns(10)
	if len(patterns) == 0 {
		t.Error("应该有查询模式")
	}

	// 按表获取查询
	queries := monitor.GetQueriesByTable("users", 10)
	if len(queries) == 0 {
		t.Error("应该有用户表查询记录")
	}

	// 按操作类型获取查询
	selectQueries := monitor.GetQueriesByOperation("SELECT", 10)
	if len(selectQueries) == 0 {
		t.Error("应该有SELECT操作记录")
	}
}

// TestSystemMonitor 测试系统监控功能
func TestSystemMonitor(t *testing.T) {
	monitor := NewMonitor(nil)
	monitor.Start()
	defer monitor.Stop()

	// 等待系统监控收集数据
	time.Sleep(100 * time.Millisecond)

	// 获取系统统计
	stats := monitor.GetSystemStats(10)
	if len(stats) == 0 {
		t.Log("系统统计可能为空，这是正常的")
	}

	// 获取最新系统状态
	latestStats := monitor.GetLatestSystemStats()
	if latestStats == nil {
		t.Log("最新系统状态可能为空，这是正常的")
	}

	// 获取系统摘要
	summary := monitor.GetSystemSummary()
	if summary == nil {
		t.Error("系统摘要不应该为空")
	}

	// 检查摘要内容
	if _, ok := summary["timestamp"]; !ok {
		t.Error("摘要应该包含时间戳")
	}
	if _, ok := summary["config"]; !ok {
		t.Error("摘要应该包含配置")
	}
}

// TestMetrics 测试指标收集功能
func TestMetrics(t *testing.T) {
	monitor := NewMonitor(nil)
	monitor.Start()
	defer monitor.Stop()

	// 记录HTTP请求指标
	monitor.RecordHTTPRequest("GET", "/api/test", "200", "TestHandler",
		100*time.Millisecond, 1024, 2048)

	// 记录数据库查询指标
	monitor.RecordDBQuery("SELECT", "test_table", "SELECT", 50*time.Millisecond)

	// 记录缓存命中/未命中
	monitor.RecordCacheHit("redis", "get")
	monitor.RecordCacheMiss("redis", "get")

	// 设置系统指标
	monitor.SetSystemMetric("active_users", "user", 100.0)

	// 获取指标管理器
	metrics := monitor.GetMetrics()
	if metrics == nil {
		t.Error("指标管理器不应该为空")
	}
}

// TestMonitorConfig 测试监控配置
func TestMonitorConfig(t *testing.T) {
	// 测试自定义配置
	config := &MonitorConfig{
		EnableMetrics:       true,
		EnableTracing:       false,
		MaxSpans:            5000,
		EnableSQLAnalysis:   true,
		MaxQueries:          5000,
		SlowThreshold:       50 * time.Millisecond,
		EnableSystemMonitor: false,
		MaxStats:            500,
		MonitorInterval:     60 * time.Second,
	}

	monitor := NewMonitor(config)
	monitor.Start()
	defer monitor.Stop()

	// 检查配置是否正确应用
	actualConfig := monitor.GetConfig()
	if actualConfig.EnableTracing {
		t.Error("链路追踪应该被禁用")
	}
	if actualConfig.EnableSystemMonitor {
		t.Error("系统监控应该被禁用")
	}
	if actualConfig.SlowThreshold != 50*time.Millisecond {
		t.Error("慢查询阈值应该正确设置")
	}

	// 检查监控状态
	if monitor.GetTracer() != nil {
		t.Error("链路追踪器应该为空")
	}
	if monitor.GetSystemMonitor() != nil {
		t.Error("系统监控器应该为空")
	}
}
