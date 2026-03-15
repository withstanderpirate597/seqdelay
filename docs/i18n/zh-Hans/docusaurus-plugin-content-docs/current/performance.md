---
sidebar_position: 8
---

# 性能

Apple M4 / arm64，纯调度引擎（无 Redis）。

| 测试 | 结果 |
|------|------|
| 100 万插入 | 225ms（440 万 tasks/sec） |
| 100 万插入+触发 | 1.2s（全部触发，零丢失） |
| 4 并发写 x 25 万 | 187ms（530 万 tasks/sec） |
| 100 万取消 | 268ms（290 ns/op） |

```bash
# 时间轮 benchmark
go test -bench=BenchmarkTimeWheel -benchmem

# 压力测试
go test -run TestStress -v -timeout 120s
```
