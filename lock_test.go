package seqdelay

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// getTestRedis returns a Redis client pointed at DB 15 (test database).
// The test is skipped when running with -short or when Redis is unavailable.
func getTestRedis(tb testing.TB) redis.UniversalClient {
	tb.Helper()
	if testing.Short() {
		tb.Skip("skipping Redis test in short mode")
	}
	client := redis.NewClient(&redis.Options{Addr: "localhost:6379", DB: 15})
	if err := client.Ping(context.Background()).Err(); err != nil {
		tb.Skipf("Redis not available: %v", err)
	}
	client.FlushDB(context.Background())
	return client
}

func TestDistLock_TryAcquire_SucceedsOnEmpty(t *testing.T) {
	client := getTestRedis(t)
	defer client.Close()

	lock := newDistLock(client, "instance-A", 500*time.Millisecond)
	if !lock.TryAcquire(context.Background()) {
		t.Fatal("TryAcquire should succeed on an empty key")
	}
}

func TestDistLock_TryAcquire_FailsWhenHeld(t *testing.T) {
	client := getTestRedis(t)
	defer client.Close()

	owner := newDistLock(client, "instance-A", 500*time.Millisecond)
	contender := newDistLock(client, "instance-B", 500*time.Millisecond)

	if !owner.TryAcquire(context.Background()) {
		t.Fatal("owner TryAcquire should succeed")
	}
	if contender.TryAcquire(context.Background()) {
		t.Fatal("contender TryAcquire should fail while lock is held")
	}
}

func TestDistLock_Renew_SucceedsForOwner(t *testing.T) {
	client := getTestRedis(t)
	defer client.Close()

	lock := newDistLock(client, "instance-A", 500*time.Millisecond)
	if !lock.TryAcquire(context.Background()) {
		t.Fatal("TryAcquire should succeed")
	}
	if !lock.Renew(context.Background()) {
		t.Fatal("Renew should succeed for the owner")
	}
}

func TestDistLock_Renew_FailsForNonOwner(t *testing.T) {
	client := getTestRedis(t)
	defer client.Close()

	owner := newDistLock(client, "instance-A", 500*time.Millisecond)
	other := newDistLock(client, "instance-B", 500*time.Millisecond)

	if !owner.TryAcquire(context.Background()) {
		t.Fatal("owner TryAcquire should succeed")
	}
	if other.Renew(context.Background()) {
		t.Fatal("Renew should fail for a non-owner")
	}
}

func TestDistLock_Release_SucceedsForOwner(t *testing.T) {
	client := getTestRedis(t)
	defer client.Close()

	lock := newDistLock(client, "instance-A", 500*time.Millisecond)
	if !lock.TryAcquire(context.Background()) {
		t.Fatal("TryAcquire should succeed")
	}
	if !lock.Release(context.Background()) {
		t.Fatal("Release should succeed for the owner")
	}
	// Lock should now be free; a contender can acquire it.
	contender := newDistLock(client, "instance-B", 500*time.Millisecond)
	if !contender.TryAcquire(context.Background()) {
		t.Fatal("contender TryAcquire should succeed after owner released the lock")
	}
}

func TestDistLock_Release_FailsForNonOwner(t *testing.T) {
	client := getTestRedis(t)
	defer client.Close()

	owner := newDistLock(client, "instance-A", 500*time.Millisecond)
	other := newDistLock(client, "instance-B", 500*time.Millisecond)

	if !owner.TryAcquire(context.Background()) {
		t.Fatal("owner TryAcquire should succeed")
	}
	if other.Release(context.Background()) {
		t.Fatal("Release should fail for a non-owner")
	}
	// Original owner should still hold the lock.
	if !owner.IsLeader(context.Background()) {
		t.Fatal("owner should still be the leader after a failed foreign release")
	}
}

func TestDistLock_IsLeader_TrueAndFalse(t *testing.T) {
	client := getTestRedis(t)
	defer client.Close()

	owner := newDistLock(client, "instance-A", 500*time.Millisecond)
	other := newDistLock(client, "instance-B", 500*time.Millisecond)

	if owner.IsLeader(context.Background()) {
		t.Fatal("IsLeader should be false before acquiring")
	}
	if !owner.TryAcquire(context.Background()) {
		t.Fatal("TryAcquire should succeed")
	}
	if !owner.IsLeader(context.Background()) {
		t.Fatal("IsLeader should be true after acquiring")
	}
	if other.IsLeader(context.Background()) {
		t.Fatal("IsLeader should be false for non-owner")
	}
}

func TestDistLock_ExpiresAfterTTL(t *testing.T) {
	client := getTestRedis(t)
	defer client.Close()

	// Use a very short TTL so the lock expires within the test.
	const shortTTL = 50 * time.Millisecond

	owner := newDistLock(client, "instance-A", shortTTL)
	if !owner.TryAcquire(context.Background()) {
		t.Fatal("owner TryAcquire should succeed")
	}

	// Wait for the lock to expire (sleep slightly longer than the TTL).
	time.Sleep(shortTTL * 3)

	if owner.IsLeader(context.Background()) {
		t.Fatal("lock should have expired; IsLeader should be false")
	}

	// A new instance should now be able to acquire the lock.
	newcomer := newDistLock(client, "instance-B", 500*time.Millisecond)
	if !newcomer.TryAcquire(context.Background()) {
		t.Fatal("newcomer TryAcquire should succeed after lock expiry")
	}
	if !newcomer.IsLeader(context.Background()) {
		t.Fatal("newcomer should be the leader after acquiring the expired lock")
	}
}
