// Distributed deployment: multiple instances share the same Redis.
// Only one instance (leader) advances the time wheel.
// All instances can Add/Pop/Finish.
//
// Run two instances in separate terminals:
//
//	INSTANCE_ID=node-1 go run .
//	INSTANCE_ID=node-2 go run .
//
// Then add tasks from either instance — they share the same Redis.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/gocronx/seqdelay"
)

func main() {
	instanceID := os.Getenv("INSTANCE_ID")
	if instanceID == "" {
		instanceID = "node-1"
	}

	q, err := seqdelay.New(
		seqdelay.WithRedis("localhost:6379"),
		seqdelay.WithInstanceID(instanceID),
		seqdelay.WithLockTTL(500*time.Millisecond),
	)
	if err != nil {
		panic(err)
	}

	q.OnExpire("jobs", func(ctx context.Context, task *seqdelay.Task) error {
		fmt.Printf("[%s] processed task %s: %s\n", instanceID, task.ID, task.Body)
		return nil
	})

	ctx := context.Background()
	q.Start(ctx)

	// Add some tasks
	for i := 0; i < 5; i++ {
		q.Add(ctx, &seqdelay.Task{
			ID:    fmt.Sprintf("%s-job-%d", instanceID, i),
			Topic: "jobs",
			Body:  []byte(fmt.Sprintf("from %s", instanceID)),
			Delay: time.Duration(i+1) * 200 * time.Millisecond,
			TTR:   5 * time.Second,
		})
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	<-sigCh

	shutdownCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	q.Shutdown(shutdownCtx)
	fmt.Printf("[%s] shutdown\n", instanceID)
}
