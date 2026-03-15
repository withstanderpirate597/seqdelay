// HTTP server mode: standalone delay queue service
//
// Usage:
//
//	go run . &
//
//	# Add a task (fires after 5 seconds)
//	curl -X POST http://localhost:9280/add \
//	  -d '{"topic":"notify","id":"msg-1","body":"hello","delay_ms":5000,"ttr_ms":30000}'
//
//	# Pop a ready task (long-poll, waits up to 10 seconds)
//	curl -X POST http://localhost:9280/pop \
//	  -d '{"topic":"notify","timeout_ms":10000}'
//
//	# Finish the task
//	curl -X POST http://localhost:9280/finish \
//	  -d '{"topic":"notify","id":"msg-1"}'
//
//	# Check stats
//	curl http://localhost:9280/stats
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

	if err := q.Start(context.Background()); err != nil {
		panic(err)
	}

	srv := seqdelay.NewServer(q, seqdelay.WithServerAddr(":9280"))

	// Graceful shutdown on SIGINT
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, os.Interrupt)
		<-sigCh
		fmt.Println("\nshutting down...")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		srv.Shutdown(ctx)
		q.Shutdown(ctx)
	}()

	fmt.Println("seqdelay HTTP server listening on :9280")
	if err := srv.ListenAndServe(); err != nil {
		fmt.Printf("server stopped: %v\n", err)
	}
}
