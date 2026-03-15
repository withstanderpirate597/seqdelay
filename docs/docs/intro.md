---
sidebar_position: 1
---

# Introduction

**seqdelay** is a high-performance delay queue for Go, powered by [seqflow](https://github.com/gocronx/seqflow)'s time wheel engine and backed by Redis for persistence and distributed coordination.

## Why seqdelay

Existing delay queue solutions (like ouqiang/delay-queue) poll Redis Sorted Sets every second. This limits precision to ~1s and puts constant load on Redis.

seqdelay uses an in-memory time wheel for scheduling (O(1) add/fire, configurable 1ms precision) and Redis only for persistence and coordination. The result: microsecond precision, millions of tasks per second, and minimal Redis load.

## Two Modes

- **Embedded** — import as a Go library, register callbacks, zero network overhead
- **HTTP Server** — deploy as a standalone service, any language can interact via HTTP

## Key Features

- Configurable precision (100us ~ 1s, default 1ms)
- Redis persistence (Standalone / Sentinel / Cluster)
- Distributed lock for multi-instance deployment
- Atomic state transitions via Lua scripts
- Full recovery from Redis on restart
- 4.4M tasks/sec scheduling throughput
