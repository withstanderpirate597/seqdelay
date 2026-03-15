---
sidebar_position: 2
---

# 快速开始

## 前置条件

- Go 1.22+
- Redis 6+

## 安装

```bash
go get github.com/gocronx/seqdelay
```

## 嵌入模式（回调）

```go
q, _ := seqdelay.New(seqdelay.WithRedis("localhost:6379"))

q.OnExpire("order-cancel", func(ctx context.Context, task *seqdelay.Task) error {
    fmt.Printf("取消订单 %s\n", task.ID)
    return nil // nil = 自动 Finish；error = TTR 后重投
})

ctx := context.Background()
q.Start(ctx)

q.Add(ctx, &seqdelay.Task{
    ID:    "order-123",
    Topic: "order-cancel",
    Body:  []byte(`{"order_id":"123"}`),
    Delay: 15 * time.Minute,
    TTR:   30 * time.Second,
})
```

## HTTP 模式

```go
q, _ := seqdelay.New(seqdelay.WithRedis("localhost:6379"))
q.Start(ctx)

srv := seqdelay.NewServer(q, seqdelay.WithServerAddr(":9280"))
srv.ListenAndServe()
```

```bash
# 添加任务
curl -X POST localhost:9280/add \
  -d '{"topic":"notify","id":"msg-1","delay_ms":5000,"ttr_ms":30000}'

# 拉取就绪任务
curl -X POST localhost:9280/pop -d '{"topic":"notify"}'

# 完成
curl -X POST localhost:9280/finish -d '{"topic":"notify","id":"msg-1"}'
```
