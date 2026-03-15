package seqdelay

import (
	"time"

	"github.com/redis/go-redis/v9"
)

// Option 配置 Queue
type Option func(*config)

type config struct {
	redisClient    redis.UniversalClient
	tickInterval   time.Duration
	wheelCapacity  uint32
	readyQueueSize int
	maxTopics      int
	popTimeout     time.Duration
	instanceID     string
	lockTTL        time.Duration
}

func defaultConfig() *config {
	return &config{
		tickInterval:   time.Millisecond,
		wheelCapacity:  4096,
		readyQueueSize: 1024,
		maxTopics:      1024,
		popTimeout:     30 * time.Second,
		lockTTL:        500 * time.Millisecond,
	}
}

// WithRedis 设置 Redis 单机连接
func WithRedis(addr string) Option {
	return func(c *config) {
		c.redisClient = redis.NewClient(&redis.Options{Addr: addr})
	}
}

// WithRedisOptions 设置带完整选项的 Redis 连接
func WithRedisOptions(opts *redis.Options) Option {
	return func(c *config) {
		c.redisClient = redis.NewClient(opts)
	}
}

// WithRedisClient 设置已有的 Redis 客户端
func WithRedisClient(client redis.UniversalClient) Option {
	return func(c *config) { c.redisClient = client }
}

// WithRedisSentinel 设置 Redis Sentinel 连接
func WithRedisSentinel(addrs []string, master string) Option {
	return func(c *config) {
		c.redisClient = redis.NewFailoverClient(&redis.FailoverOptions{
			SentinelAddrs: addrs,
			MasterName:    master,
		})
	}
}

// WithRedisCluster 设置 Redis Cluster 连接
func WithRedisCluster(addrs []string) Option {
	return func(c *config) {
		c.redisClient = redis.NewClusterClient(&redis.ClusterOptions{Addrs: addrs})
	}
}

// WithTickInterval 设置时间轮精度（默认 1ms）
func WithTickInterval(d time.Duration) Option {
	return func(c *config) { c.tickInterval = d }
}

// WithWheelCapacity 设置时间轮容量（默认 4096，必须 2 的幂）
func WithWheelCapacity(n uint32) Option {
	return func(c *config) { c.wheelCapacity = n }
}

// WithReadyQueueSize 设置每个 topic 就绪队列缓冲大小（默认 1024）
func WithReadyQueueSize(n int) Option {
	return func(c *config) { c.readyQueueSize = n }
}

// WithMaxTopics 设置最大 topic 数（默认 1024）
func WithMaxTopics(n int) Option {
	return func(c *config) { c.maxTopics = n }
}

// WithPopTimeout 设置 HTTP Pop 默认超时（默认 30s）
func WithPopTimeout(d time.Duration) Option {
	return func(c *config) { c.popTimeout = d }
}

// WithInstanceID 设置实例 ID（分布式锁用，默认自动生成）
func WithInstanceID(id string) Option {
	return func(c *config) { c.instanceID = id }
}

// WithLockTTL 设置分布式锁 TTL（默认 500ms，心跳续期）
func WithLockTTL(d time.Duration) Option {
	return func(c *config) { c.lockTTL = d }
}
