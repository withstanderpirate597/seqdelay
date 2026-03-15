---
sidebar_position: 8
---

# Performance

## Time Wheel Benchmarks

Apple M4 / arm64. Pure scheduling engine, no Redis.

| Test | Result |
|------|--------|
| 1M Add | 225ms (4.4M tasks/sec, 225 ns/op) |
| 1M Add + Fire | 1.2s (all fired, zero loss) |
| 4 writers x 250K | 187ms (5.3M tasks/sec) |
| 1M Cancel | 268ms (290 ns/op) |

## vs Redis Sorted Set

| Operation | seqdelay (time wheel) | Redis ZRANGEBYSCORE |
|-----------|----------------------|---------------------|
| Add | 225 ns (local memory) | ~0.5 ms (network RTT) |
| Fire check | O(1) per tick | O(log N) per poll |
| Precision | 1ms configurable | ~1s (poll interval) |
| Redis load | On state change only | Every poll cycle |

## Running Benchmarks

```bash
# Time wheel only (no Redis needed)
go test -bench=BenchmarkTimeWheel -benchmem -count=1

# Stress test: 1M tasks
go test -run TestStress -v -timeout 120s

# Full lifecycle (needs Redis)
go test -bench=BenchmarkQueue -benchmem -timeout 60s

# All benchmarks
go test -bench=. -benchmem -count=1 -timeout 120s
```
