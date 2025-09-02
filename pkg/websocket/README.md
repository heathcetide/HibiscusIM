# WebSocket 模块

高性能WebSocket模块，支持几万~百万并发连接（取决于机器与部署拓扑）。

## 特性

- 🚀 支持10万+并发连接（可扩展至百万级）
- 🔄 自动心跳检测（支持全局心跳以减少 goroutine/timer 开销）
- 👥 用户和组管理
- 📊 实时监控
- ⚙️ 灵活配置

## 快速开始

```go
// 创建Hub
hub := websocket.NewHub(nil)
defer hub.Close()

// 创建处理器
handler := websocket.NewHandler(hub)

// 设置路由（推荐）
websocket.RegisterRoutes(r, handler)
```

## 配置

支持从环境变量加载：

```bash
export WEBSOCKET_MAX_CONNECTIONS=100000
export WEBSOCKET_HEARTBEAT_INTERVAL=30           # 秒
export WEBSOCKET_CONNECTION_TIMEOUT=60           # 秒
export WEBSOCKET_MESSAGE_BUFFER_SIZE=256
export WEBSOCKET_MESSAGE_QUEUE_SIZE=1000
export WEBSOCKET_ENABLE_COMPRESSION=1            # 1/true 启用
export WEBSOCKET_SHARD_COUNT=16
export WEBSOCKET_BROADCAST_WORKERS=32
export WEBSOCKET_DROP_ON_FULL=1                  # 背压策略：满则丢弃
export WEBSOCKET_CLOSE_ON_BACKPRESSURE=0         # 背压时是否断开慢连接
export WEBSOCKET_SEND_TIMEOUT_MS=50              # 非丢弃模式下的发送超时(ms)
export WEBSOCKET_COMPRESSION_LEVEL=-2            # -2..9
export WEBSOCKET_READ_BUFFER_SIZE=1024
export WEBSOCKET_WRITE_BUFFER_SIZE=1024
export WEBSOCKET_MAX_MESSAGE_SIZE=512
# 全局心跳（百万连接推荐开启）
export WEBSOCKET_ENABLE_GLOBAL_PING=1
export WEBSOCKET_PING_WORKERS=16
```

也可通过代码合并配置：

```go
cfg := websocket.MergeConfig(
    websocket.DefaultConfig(),
    &websocket.Config{ MaxConnections: 200000, ShardCount: 32, EnableGlobalPing: true, PingWorkerCount: 16 },
)
hub := websocket.NewHub(cfg)
```

## 消息格式

```json
{
  "type": "chat",
  "data": "Hello World",
  "timestamp": 1234567890,
  "from": "user1",
  "to": "user2"
}
```

## API接口

- `GET /ws` - WebSocket连接
- `GET /ws/stats` - 连接统计
- `GET /ws/health` - 健康检查
- `POST /ws/message` - 发送消息
- `POST /ws/broadcast` - 广播消息

## 性能与调优建议

- 应用级
  - 分片与并发：`ShardCount`、`BroadcastWorkerCount`、`PingWorkerCount` 可按 CPU 核心数×(2~4) 起步。
  - 慢消费者处理：`DROP_ON_FULL=1` 保护整体；或 `CLOSE_ON_BACKPRESSURE=1` 直接断开慢连接。
  - 启用 `EnableGlobalPing=1`，避免每连接 ticker；Pong 回调更新 `LastPing`。
  - 控制消息大小与频率，尽可能批量发送与复用编码缓存。
- 系统级
  - 文件句柄：`ulimit -n` 提升至百万级；
  - 内核参数：`somaxconn`、`rmem_max/wmem_max`、`tcp_max_syn_backlog`、`ip_local_port_range`；
  - 网络与 LB：四层直连/DSR，或在多节点间利用外部总线做跨节点广播。
- 压测
  - 使用 `autocannon --ws`、自研 ws 压测器或 go+gorilla ws 客户端批量建连，逐步升压并监控 CPU、内存、FD、GC、goroutine 数量与延迟分布。




// 1) 选择存储（默认内存；企业推荐 Redis store）
store := memory.NewStore() // 或 redis.NewStoreWithOptions(client, limiter.StoreOptions{Prefix: "rl:"})

// 2) 定义配置（支持路由覆盖速率）
cfg := middleware.RateLimiterConfig{
  Rate:         "200-M",
  PerRouteRates: map[string]string{
    "/api/v1/heavy":  "10-S",
    "/api/v1/normal": "100-S",
  },
  Identifier:     "ip+route",                 // ip | user | header | ip+route
  HeaderName:     "X-Client-ID",              // 当 Identifier=header 时生效
  WhitelistCIDRs: []string{"10.0.0.0/8"},
  SkipPaths:      []string{"/health", "/metrics", "/static/"},
  AddHeaders:     true,
  DenyStatus:     429,
  DenyMessage:    "Too Many Requests",
}

// 3) 构造实例并挂载中间件
rl := middleware.NewRateLimiter(cfg, store)
r.Use(rl.Middleware())

// 4) 运行时热更新（不重启）
rl.UpdateConfig(cfg2)

// 1) Redis store 工厂
type RedisStoreFactory struct{ client *redis.Client }
func (f *RedisStoreFactory) Create() limiter.Store {
  return redisstore.NewStoreWithOptions(f.client, limiter.StoreOptions{Prefix: "rl:"})
}

// 2) 指标观察者
type PromObserver struct{}
func (p *PromObserver) OnAllow(route, key string) { /* promCounterAllow.With(...).Inc() */ }
func (p *PromObserver) OnDeny(route, key string)  { /* promCounterDeny.With(...).Inc() */ }

// 3) 构造与挂载
rl := middleware.NewRateLimiter(cfg, nil).
  WithStoreFactory(&RedisStoreFactory{client: redisCli}).
  WithObserver(&PromObserver{})
r.Use(rl.Middleware())