# 监控系统使用指南

## 概述

这是一个完整的单体应用监控系统，包含以下功能：

- **指标收集**: 基于Prometheus的HTTP、数据库、缓存等指标收集
- **链路追踪**: 分布式链路追踪，支持请求链路分析
- **SQL分析**: 慢查询检测、查询模式分析、性能诊断
- **系统监控**: CPU、内存、磁盘、网络等系统资源监控

## 快速开始

### 1. 创建监控器

```go
package main

import (
    "time"
    "github.com/your-project/pkg/metrics"
)

func main() {
    // 使用默认配置创建监控器
    monitor := metrics.NewMonitor(nil)
    
    // 或者使用自定义配置
    config := &metrics.MonitorConfig{
        EnableMetrics:      true,
        EnableTracing:      true,
        MaxSpans:           10000,
        EnableSQLAnalysis:  true,
        MaxQueries:         10000,
        SlowThreshold:      100 * time.Millisecond,
        EnableSystemMonitor: true,
        MaxStats:           1000,
        MonitorInterval:     30 * time.Second,
    }
    
    monitor = metrics.NewMonitor(config)
    
    // 启动监控
    monitor.Start()
    defer monitor.Stop()
}
```

### 2. 在Gin中使用监控中间件

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/your-project/pkg/metrics"
)

func main() {
    r := gin.Default()
    
    // 创建监控器
    monitor := metrics.NewMonitor(nil)
    monitor.Start()
    
    // 使用监控中间件
    r.Use(metrics.MonitorMiddleware(monitor))
    
    // 注册监控API路由
    monitorAPI := metrics.NewMonitorAPI(monitor)
    apiGroup := r.Group("/api/monitor")
    monitorAPI.RegisterRoutes(apiGroup)
    
    r.Run(":8080")
}
```

### 3. 手动记录指标

```go
// 记录HTTP请求指标
monitor.RecordHTTPRequest("GET", "/api/users", "200", "GetUsers", 
    time.Since(start), requestSize, responseSize)

// 记录数据库查询指标
monitor.RecordDBQuery("SELECT", "users", "SELECT", time.Since(start))

// 记录缓存命中/未命中
monitor.RecordCacheHit("redis", "get")
monitor.RecordCacheMiss("redis", "get")

// 记录SQL查询（自动分析慢查询）
monitor.RecordSQLQuery(ctx, "SELECT * FROM users WHERE id = ?", 
    []interface{}{123}, "users", "SELECT", duration, rowsAffected, err)
```

### 4. 链路追踪

```go
// 在业务逻辑中使用链路追踪
func (h *UserHandler) GetUser(c *gin.Context) {
    // 开始业务跨度
    ctx, span := monitor.StartSpan(c.Request.Context(), "GetUser",
        WithTags(map[string]string{
            "user_id": c.Param("id"),
            "action":  "get_user",
        }),
    )
    defer monitor.EndSpan(span, nil)
    
    // 添加事件
    span.AddEvent("user_found", map[string]interface{}{
        "user_id": userID,
        "found":   true,
    })
    
    // 业务逻辑...
}
```

## 监控数据访问

### 1. 系统概览

```bash
GET /api/monitor/overview
```

返回系统整体状态，包括：
- 系统资源使用情况
- SQL查询统计
- 链路追踪统计
- 监控配置信息

### 2. 系统监控

```bash
# 获取系统统计历史
GET /api/monitor/system?limit=100

# 获取最新系统状态
GET /api/monitor/system/latest
```

### 3. SQL分析

```bash
# 获取慢查询列表
GET /api/monitor/sql/slow?limit=50

# 获取查询模式分析
GET /api/monitor/sql/patterns?limit=50

# 获取SQL统计信息
GET /api/monitor/sql/stats

# 按表查询
GET /api/monitor/sql/table/users?limit=50

# 按操作类型查询
GET /api/monitor/sql/operation/SELECT?limit=50
```

### 4. 链路追踪

```bash
# 获取所有追踪
GET /api/monitor/traces

# 获取特定追踪详情
GET /api/monitor/traces/trace_1234567890
```

## 慢SQL分析功能

### 自动检测

系统会自动检测执行时间超过阈值的SQL查询：

```go
// 设置慢查询阈值
config := &metrics.MonitorConfig{
    SlowThreshold: 100 * time.Millisecond, // 100ms
}

// 记录查询时会自动分析
monitor.RecordSQLQuery(ctx, sql, params, table, operation, duration, rowsAffected, err)
```

### 查询模式分析

系统会自动分析SQL查询模式，识别：
- 高频查询
- 耗时较长的查询模式
- 表访问频率
- 操作类型分布

### 性能诊断

通过分析查询模式，可以：
- 识别需要优化的查询
- 发现索引缺失
- 监控查询性能趋势
- 预警性能问题

## 系统资源监控

### 监控指标

- **CPU**: 使用率、负载、温度
- **内存**: 使用量、可用量、交换分区
- **磁盘**: 使用量、IO统计、读写性能
- **网络**: 流量统计、错误统计、接口状态
- **进程**: 资源使用、状态信息
- **运行时**: Goroutine数量、内存分配、GC统计

### 数据收集

```go
// 系统监控器会自动收集系统指标
monitor := metrics.NewMonitor(&metrics.MonitorConfig{
    EnableSystemMonitor: true,
    MonitorInterval:     30 * time.Second, // 每30秒收集一次
})

monitor.Start()
```

## 配置选项

### 监控配置

```go
type MonitorConfig struct {
    // 指标收集
    EnableMetrics bool
    
    // 链路追踪
    EnableTracing bool
    MaxSpans      int
    
    // SQL分析
    EnableSQLAnalysis bool
    MaxQueries        int
    SlowThreshold     time.Duration
    
    // 系统监控
    EnableSystemMonitor bool
    MaxStats            int
    MonitorInterval     time.Duration
}
```

### 性能调优

```go
// 根据系统规模调整配置
config := &metrics.MonitorConfig{
    // 高并发系统
    MaxSpans:      50000,  // 增加跨度存储
    MaxQueries:    50000,  // 增加查询记录
    MaxStats:      5000,   // 增加系统统计
    
    // 实时监控
    MonitorInterval: 10 * time.Second, // 更频繁的收集
    
    // 性能敏感场景
    SlowThreshold: 50 * time.Millisecond, // 更严格的慢查询阈值
}
```

## 注意事项

1. **内存使用**: 监控数据会占用内存，根据系统规模调整存储限制
2. **性能影响**: 监控会带来少量性能开销，生产环境建议适当调整收集频率
3. **平台兼容**: 某些系统监控功能在不同操作系统上可能不可用
4. **数据清理**: 系统会自动清理旧数据，避免内存泄漏

## 故障排查

### 常见问题

1. **监控数据为空**: 检查监控器是否启动，配置是否正确
2. **内存占用过高**: 减少MaxSpans、MaxQueries、MaxStats等配置值
3. **性能影响**: 增加MonitorInterval，减少监控频率
4. **平台兼容**: 某些监控功能在特定平台不可用，系统会自动降级

### 调试模式

```go
// 启用详细日志
config := &metrics.MonitorConfig{
    EnableMetrics:      true,
    EnableTracing:      true,
    EnableSQLAnalysis:  true,
    EnableSystemMonitor: true,
}

// 检查监控状态
if monitor.IsEnabled() {
    log.Println("监控系统已启用")
}

// 获取监控摘要
summary := monitor.GetSystemSummary()
log.Printf("监控摘要: %+v", summary)
```
