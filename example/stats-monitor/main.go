// Stats monitor: add tasks and watch queue statistics via HTTP.
//
// Usage:
//
//	go run . &
//	watch -n 1 'curl -s localhost:9280/stats | python3 -m json.tool'
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
	q, err := seqdelay.New(
		seqdelay.WithRedis("localhost:6379"),
		seqdelay.WithTickInterval(time.Millisecond),
	)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	q.Start(ctx)

	srv := seqdelay.NewServer(q, seqdelay.WithServerAddr(":9280"))
	go srv.ListenAndServe()

	// Add tasks across multiple topics with varying delays
	topics := []string{"orders", "payments", "notifications"}
	for i := 0; i < 30; i++ {
		topic := topics[i%len(topics)]
		q.Add(ctx, &seqdelay.Task{
			ID:    fmt.Sprintf("task-%d", i),
			Topic: topic,
			Body:  []byte(fmt.Sprintf(`{"seq":%d}`, i)),
			Delay: time.Duration(i+1) * 200 * time.Millisecond,
			TTR:   10 * time.Second,
		})
	}
	fmt.Println("added 30 tasks across 3 topics")
	fmt.Println("monitor stats: curl localhost:9280/stats")

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	<-sigCh

	shutdownCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	srv.Shutdown(shutdownCtx)
	q.Shutdown(shutdownCtx)
}
