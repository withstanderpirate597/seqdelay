package seqdelay

import "time"

// TaskState 表示任务生命周期状态
type TaskState int

const (
	StateDelayed   TaskState = iota // 在时间轮中等待
	StateReady                      // 在就绪队列中
	StateActive                     // 已被消费，TTR 倒计时中
	StateFinished                   // 已完成
	StateCancelled                  // 已取消
)

func (s TaskState) String() string {
	switch s {
	case StateDelayed:
		return "delayed"
	case StateReady:
		return "ready"
	case StateActive:
		return "active"
	case StateFinished:
		return "finished"
	case StateCancelled:
		return "cancelled"
	default:
		return "unknown"
	}
}

// Task 表示一个延迟任务
type Task struct {
	ID         string        `msgpack:"id" json:"id"`
	Topic      string        `msgpack:"topic" json:"topic"`
	Body       []byte        `msgpack:"body" json:"body"`
	Delay      time.Duration `msgpack:"delay" json:"delay_ms"`
	TTR        time.Duration `msgpack:"ttr" json:"ttr_ms"`
	MaxRetries int           `msgpack:"max_retries" json:"max_retries"`
	State      TaskState     `msgpack:"state" json:"state"`
	Retries    int           `msgpack:"retries" json:"retries"`
	CreatedAt  time.Time     `msgpack:"created_at" json:"created_at"`
	ActiveAt   time.Time     `msgpack:"active_at" json:"active_at"`
}

// Validate 校验必填字段
func (t *Task) Validate() error {
	if t.ID == "" || t.Topic == "" {
		return ErrInvalidTask
	}
	if t.Delay <= 0 {
		return ErrInvalidDelay
	}
	if t.TTR <= 0 {
		t.TTR = 30 * time.Second
	}
	return nil
}

// taskKey 返回任务的 Redis key（使用 {topic} hash tag 保证 Cluster 兼容）
func taskKey(topic, id string) string {
	return "seqdelay:{" + topic + "}:task:" + id
}

// readyKey 返回 topic 的就绪队列 Redis key
func readyKey(topic string) string {
	return "seqdelay:{" + topic + "}:ready"
}

// indexKey 返回 topic 的任务索引 Redis key
func indexKey(topic string) string {
	return "seqdelay:{" + topic + "}:index"
}

// lockKey 分布式锁 key
const lockKey = "seqdelay:lock:tick"
