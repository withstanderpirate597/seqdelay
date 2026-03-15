---
sidebar_position: 2
---

# Getting Started

## Prerequisites

- Go 1.22+
- Redis 6+

## Install

```bash
go get github.com/gocronx/seqdelay
```

## Embedded Mode (Callback)

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/gocronx/seqdelay"
)

func main() {
    q, _ := seqdelay.New(seqdelay.WithRedis("localhost:6379"))

    q.OnExpire("order-cancel", func(ctx context.Context, task *seqdelay.Task) error {
        fmt.Printf("cancelling order %s\n", task.ID)
        return nil // nil = auto Finish; error = redeliver after TTR
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

    // User pays → cancel the auto-cancel task
    // q.Cancel(ctx, "order-cancel", "order-123")
}
```

## HTTP Mode (Pull)

```go
q, _ := seqdelay.New(seqdelay.WithRedis("localhost:6379"))
q.Start(ctx)

srv := seqdelay.NewServer(q, seqdelay.WithServerAddr(":9280"))
srv.ListenAndServe()
```

```bash
# Add a task (fires after 5 seconds)
curl -X POST localhost:9280/add \
  -d '{"topic":"notify","id":"msg-1","delay_ms":5000,"ttr_ms":30000}'

# Pop a ready task
curl -X POST localhost:9280/pop -d '{"topic":"notify"}'

# Finish
curl -X POST localhost:9280/finish -d '{"topic":"notify","id":"msg-1"}'
```
