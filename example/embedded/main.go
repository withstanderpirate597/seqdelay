// Embedded mode: callback-driven delay queue within your Go process
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/gocronx/seqdelay"
)

func main() {
	q, err := seqdelay.New(
		seqdelay.WithRedis("localhost:6379"),
		seqdelay.WithTickInterval(time.Millisecond),
		seqdelay.WithWheelCapacity(4096),
	)
	if err != nil {
		panic(err)
	}

	// Register callback — fires when task delay expires
	q.OnExpire("order-timeout", func(ctx context.Context, task *seqdelay.Task) error {
		fmt.Printf("order expired: id=%s body=%s\n", task.ID, task.Body)
		return nil // nil = auto Finish; error = redeliver after TTR
	})

	ctx := context.Background()
	if err := q.Start(ctx); err != nil {
		panic(err)
	}

	// Add some delayed tasks
	for i := 0; i < 5; i++ {
		q.Add(ctx, &seqdelay.Task{
			ID:    fmt.Sprintf("order-%d", i),
			Topic: "order-timeout",
			Body:  []byte(fmt.Sprintf(`{"orderId":%d}`, i)),
			Delay: time.Duration(i+1) * 100 * time.Millisecond,
			TTR:   5 * time.Second,
		})
	}

	// Wait for all tasks to fire
	time.Sleep(2 * time.Second)

	shutdownCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	q.Shutdown(shutdownCtx)

	fmt.Println("done")
}
