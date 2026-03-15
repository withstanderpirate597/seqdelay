// Batch add example: add many tasks with different delays.
// Demonstrates the time wheel handling tasks across multiple slots.
package main

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/gocronx/seqdelay"
)

func main() {
	q, err := seqdelay.New(
		seqdelay.WithRedis("localhost:6379"),
		seqdelay.WithTickInterval(time.Millisecond),
	)
	if err != nil {
		panic(err)
	}

	var processed atomic.Int64

	q.OnExpire("batch", func(ctx context.Context, task *seqdelay.Task) error {
		n := processed.Add(1)
		if n%100 == 0 || n == 1000 {
			fmt.Printf("processed %d tasks\n", n)
		}
		return nil
	})

	ctx := context.Background()
	q.Start(ctx)

	// Add 1000 tasks with delays from 10ms to 1000ms
	start := time.Now()
	for i := 0; i < 1000; i++ {
		q.Add(ctx, &seqdelay.Task{
			ID:    fmt.Sprintf("item-%d", i),
			Topic: "batch",
			Body:  []byte("{}"),
			Delay: time.Duration(i+1) * time.Millisecond,
			TTR:   5 * time.Second,
		})
	}
	fmt.Printf("added 1000 tasks in %v\n", time.Since(start))

	// Wait for all to fire
	time.Sleep(3 * time.Second)

	shutdownCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	q.Shutdown(shutdownCtx)

	fmt.Printf("total processed: %d\n", processed.Load())
}
