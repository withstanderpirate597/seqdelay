// Cancel example: add a task, then cancel it before it fires.
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
	)
	if err != nil {
		panic(err)
	}

	q.OnExpire("reminders", func(ctx context.Context, task *seqdelay.Task) error {
		fmt.Printf("UNEXPECTED: task %s fired — should have been cancelled!\n", task.ID)
		return nil
	})

	ctx := context.Background()
	q.Start(ctx)

	// Add a task with 2-second delay
	q.Add(ctx, &seqdelay.Task{
		ID:    "reminder-1",
		Topic: "reminders",
		Body:  []byte("meeting at 3pm"),
		Delay: 2 * time.Second,
		TTR:   5 * time.Second,
	})
	fmt.Println("added reminder-1 (fires in 2s)")

	// Check state
	task, _ := q.Get(ctx, "reminders", "reminder-1")
	fmt.Printf("state before cancel: %s\n", task.State)

	// Cancel after 500ms
	time.Sleep(500 * time.Millisecond)
	err = q.Cancel(ctx, "reminders", "reminder-1")
	fmt.Printf("cancelled: err=%v\n", err)

	// Check state after cancel
	task, _ = q.Get(ctx, "reminders", "reminder-1")
	if task != nil {
		fmt.Printf("state after cancel: %s\n", task.State)
	}

	// Wait past the original delay — callback should NOT fire
	time.Sleep(3 * time.Second)
	fmt.Println("done — callback did not fire (correctly cancelled)")

	shutdownCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	q.Shutdown(shutdownCtx)
}
