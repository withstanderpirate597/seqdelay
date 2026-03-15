---
sidebar_position: 4
---

# Go SDK API

## 创建队列

```go
q, err := seqdelay.New(opts ...Option) (*Queue, error)
```

### 选项

| 选项 | 默认值 | 说明 |
|------|--------|------|
| `WithRedis(addr)` | 必填 | Redis 单机 |
| `WithRedisSentinel(addrs, master)` | — | Redis Sentinel |
| `WithRedisCluster(addrs)` | — | Redis Cluster |
| `WithTickInterval(d)` | 1ms | 时间轮精度 |
| `WithWheelCapacity(n)` | 4096 | 时间轮容量 |
| `WithPopTimeout(d)` | 30s | HTTP Pop 默认超时 |
| `WithLockTTL(d)` | 500ms | 分布式锁 TTL |

### 方法

| 方法 | 说明 |
|------|------|
| `Start(ctx)` | 启动（恢复 + 时间轮 + 选主） |
| `Add(ctx, *Task)` | 添加延迟任务 |
| `Pop(ctx, topic)` | 拉取就绪任务 |
| `Finish(ctx, topic, id)` | 确认完成 |
| `Cancel(ctx, topic, id)` | 取消任务 |
| `Get(ctx, topic, id)` | 查询状态 |
| `OnExpire(topic, fn)` | 注册回调 |
| `Shutdown(ctx)` | 优雅关闭 |
