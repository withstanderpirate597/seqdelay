---
sidebar_position: 4
---

# Go SDK API

## Creating a Queue

```go
q, err := seqdelay.New(opts ...Option) (*Queue, error)
```

### Options

| Option | Default | Description |
|--------|---------|-------------|
| `WithRedis(addr)` | required | Redis standalone |
| `WithRedisOptions(opts)` | — | Redis with full options (password, TLS, DB) |
| `WithRedisClient(client)` | — | Pre-configured Redis client |
| `WithRedisSentinel(addrs, master)` | — | Redis Sentinel |
| `WithRedisCluster(addrs)` | — | Redis Cluster |
| `WithTickInterval(d)` | 1ms | Time wheel precision |
| `WithWheelCapacity(n)` | 4096 | Wheel slot count (power of 2) |
| `WithReadyQueueSize(n)` | 1024 | Per-topic channel buffer |
| `WithMaxTopics(n)` | 1024 | Max topic count |
| `WithPopTimeout(d)` | 30s | Default HTTP pop timeout |
| `WithInstanceID(id)` | auto | Instance ID for distributed lock |
| `WithLockTTL(d)` | 500ms | Distributed lock TTL |

## Methods

### Start

```go
func (q *Queue) Start(ctx context.Context) error
```

Recovers tasks from Redis, starts the time wheel, begins leader election.

### Add

```go
func (q *Queue) Add(ctx context.Context, task *Task) error
```

Add a delayed task. `Task.ID` + `Task.Topic` must be unique. Returns `ErrDuplicateTask` on conflict.

### Pop

```go
func (q *Queue) Pop(ctx context.Context, topic string) (*Task, error)
```

Pull a ready task. Blocks until a task is available. Sets task state to ACTIVE and starts TTR countdown.

### Finish

```go
func (q *Queue) Finish(ctx context.Context, topic, id string) error
```

Mark a task as completed. Validates state is ACTIVE. Returns `ErrInvalidState` otherwise.

### Cancel

```go
func (q *Queue) Cancel(ctx context.Context, topic, id string) error
```

Cancel a task from any state.

### Get

```go
func (q *Queue) Get(ctx context.Context, topic, id string) (*Task, error)
```

Query task details and current state.

### OnExpire

```go
func (q *Queue) OnExpire(topic string, fn func(context.Context, *Task) error) error
```

Register a callback for a topic. Mutually exclusive with Pop on the same topic.

- `return nil` — auto Finish
- `return error` — no Finish, TTR will redeliver

### Shutdown

```go
func (q *Queue) Shutdown(ctx context.Context) error
```

Graceful shutdown. Drains the time wheel and stops all callback goroutines.

## Task Struct

```go
type Task struct {
    ID         string
    Topic      string
    Body       []byte
    Delay      time.Duration
    TTR        time.Duration
    MaxRetries int           // 0 = unlimited
}
```

## Errors

| Error | Description |
|-------|-------------|
| `ErrDuplicateTask` | Task with same topic:id already exists |
| `ErrTaskNotFound` | Task not found |
| `ErrInvalidState` | Invalid state for this operation |
| `ErrTopicConflict` | Topic already has a callback |
| `ErrTooManyTopics` | Max topic count exceeded |
| `ErrQueueFull` | Ready queue buffer full |
| `ErrClosed` | Queue is closed |
| `ErrRedisRequired` | Redis connection not provided |
| `ErrInvalidTask` | Missing ID or Topic |
| `ErrInvalidDelay` | Delay must be positive |
