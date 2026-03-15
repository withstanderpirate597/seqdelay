---
sidebar_position: 1
---

# 简介

**seqdelay** 是一个高性能延迟队列，基于 [seqflow](https://github.com/gocronx/seqflow) 时间轮引擎 + Redis 持久化和分布式协调。

## 为什么用 seqdelay

现有延迟队列方案每秒轮询 Redis Sorted Set，精度限制在 ~1s，且持续给 Redis 施加负载。

seqdelay 使用内存时间轮调度（O(1) 添加/触发，可配 1ms 精度），Redis 仅用于持久化和协调。

## 两种模式

- **嵌入模式** — 作为 Go 库导入，注册回调，零网络开销
- **HTTP 模式** — 独立部署为服务，任何语言通过 HTTP 交互
