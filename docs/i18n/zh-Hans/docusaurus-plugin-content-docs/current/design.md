---
sidebar_position: 9
---

# 设计

## 架构

```mermaid
flowchart TD
    subgraph seqdelay
        Q[队列核心]
        TW[时间轮]
        ST[Redis Store]
        RQ[就绪队列]
        LK[分布式锁]
        SRV[HTTP Server]
    end
    SF[seqflow Disruptor] -->|tick 时钟| TW
    TW -->|onFire| ST
    ST <-->|Lua 脚本| R[(Redis)]
    RQ <-->|RPUSH/BLPOP| R
    LK <-->|SetNX/续期| R
    Q --> TW & ST & RQ & LK
    SRV --> Q
```

## 时间轮

seqflow Disruptor 驱动 tick 时钟，轮盘槽位是独立数组，由 handler goroutine 独占。

## Lua 脚本

4 个原子脚本：add、finish、ready、cancel。乐观 CAS 模式。

## 就绪队列

Redis List 为持久化数据源。嵌入模式额外有 drain goroutine 拉取到本地 channel。
