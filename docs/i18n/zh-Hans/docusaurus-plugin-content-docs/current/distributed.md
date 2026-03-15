---
sidebar_position: 7
---

# 分布式部署

多个实例连接同一个 Redis，只有一个推进时间轮（leader），所有实例都可以处理 HTTP 请求。

```mermaid
graph TB
    subgraph 实例集群
        A["实例 A - leader<br/>时间轮 + HTTP"]
        B["实例 B<br/>待命 + HTTP"]
        C["实例 C<br/>待命 + HTTP"]
    end
    A & B & C <--> R[(Redis)]
    style A fill:#2563eb,color:#fff,stroke:#1d4ed8
    style B fill:#64748b,color:#fff,stroke:#475569
    style C fill:#64748b,color:#fff,stroke:#475569
    style R fill:#dc382d,color:#fff,stroke:#b91c1c
```

Leader 宕机 → 锁 500ms 过期 → standby 自动接管。

## Redis 模式

```go
seqdelay.WithRedis("localhost:6379")                         // 单机
seqdelay.WithRedisSentinel(addrs, "mymaster")                // Sentinel
seqdelay.WithRedisCluster([]string{"n1:6379", "n2:6379"})    // Cluster
```
