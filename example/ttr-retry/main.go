// TTR retry example: task is redelivered if not Finished within TTR timeout.
//
// Simulates a consumer that fails the first 2 attempts, then succeeds.
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

	var attempts atomic.Int32

	q.OnExpire("payment", func(ctx context.Context, task *seqdelay.Task) error {
		n := attempts.Add(1)
		if n <= 2 {
			fmt.Printf("attempt %d: processing %s — FAILED (will retry after TTR)\n", n, task.ID)
			return fmt.Errorf("simulated failure")
		}
		fmt.Printf("attempt %d: processing %s — SUCCESS\n", n, task.ID)
		return nil // auto Finish
	})

	ctx := context.Background()
	q.Start(ctx)

	q.Add(ctx, &seqdelay.Task{
		ID:         "pay-001",
		Topic:      "payment",
		Body:       []byte(`{"amount":99.99}`),
		Delay:      100 * time.Millisecond,
		TTR:        500 * time.Millisecond, // short TTR for demo
		MaxRetries: 5,
	})

	// Wait for retries to complete
	time.Sleep(5 * time.Second)

	shutdownCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	q.Shutdown(shutdownCtx)

	fmt.Printf("total attempts: %d\n", attempts.Load())
}
