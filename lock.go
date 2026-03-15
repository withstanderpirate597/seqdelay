package seqdelay

import (
	"context"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

// renewScript checks that the caller owns the lock, then refreshes its TTL.
// KEYS[1] — lock key
// ARGV[1] — instanceID
// ARGV[2] — TTL in milliseconds
var renewScript = redis.NewScript(`
if redis.call('GET', KEYS[1]) == ARGV[1] then
    redis.call('PEXPIRE', KEYS[1], ARGV[2])
    return 1
end
return 0
`)

// releaseScript checks that the caller owns the lock, then deletes it.
// KEYS[1] — lock key
// ARGV[1] — instanceID
var releaseScript = redis.NewScript(`
if redis.call('GET', KEYS[1]) == ARGV[1] then
    redis.call('DEL', KEYS[1])
    return 1
end
return 0
`)

// distLock is a Redis-based distributed lock that ties ownership to an
// instanceID so that only the holder can renew or release it.
type distLock struct {
	client     redis.UniversalClient
	key        string        // e.g. "seqdelay:lock:tick"
	instanceID string        // unique identifier for this process/instance
	ttl        time.Duration // how long the lock is valid without renewal
}

// newDistLock creates a distLock. ttl must be positive; callers typically
// pass config.lockTTL (default 500 ms).
func newDistLock(client redis.UniversalClient, instanceID string, ttl time.Duration) *distLock {
	if ttl <= 0 {
		ttl = 500 * time.Millisecond
	}
	return &distLock{
		client:     client,
		key:        lockKey,
		instanceID: instanceID,
		ttl:        ttl,
	}
}

// TryAcquire attempts to acquire the lock with SET NX PX.
// Returns true if this instance became the lock holder.
func (l *distLock) TryAcquire(ctx context.Context) bool {
	ok, err := l.client.SetNX(ctx, l.key, l.instanceID, l.ttl).Result()
	return err == nil && ok
}

// Renew refreshes the lock TTL if this instance still holds it.
// Returns true on a successful renewal; false if the lock has expired or
// is held by a different instance.
func (l *distLock) Renew(ctx context.Context) bool {
	ttlMs := strconv.FormatInt(l.ttl.Milliseconds(), 10)
	res, err := renewScript.Run(ctx, l.client, []string{l.key}, l.instanceID, ttlMs).Int()
	return err == nil && res == 1
}

// Release deletes the lock only if this instance is the current holder.
// Returns true when the lock was released; false if it was already gone
// or belongs to another instance.
func (l *distLock) Release(ctx context.Context) bool {
	res, err := releaseScript.Run(ctx, l.client, []string{l.key}, l.instanceID).Int()
	return err == nil && res == 1
}

// IsLeader returns true when this instance currently holds the lock.
func (l *distLock) IsLeader(ctx context.Context) bool {
	val, err := l.client.Get(ctx, l.key).Result()
	return err == nil && val == l.instanceID
}
