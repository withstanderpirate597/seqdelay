package seqdelay

import (
	"context"
	"sync"
	"time"
)

// callbackEntry holds a per-topic drain callback and the cancel function that
// stops the associated drain goroutine.
type callbackEntry struct {
	fn     func(context.Context, *Task) error
	cancel context.CancelFunc
	done   chan struct{} // closed when drain goroutine exits
}

// readyQueue manages per-topic drain goroutines that bridge the Redis ready
// list to user-provided callback functions.
type readyQueue struct {
	mu        sync.RWMutex
	callbacks map[string]callbackEntry // topic → callback + cancel
	maxTopics int
	store     *store
}

// newReadyQueue creates a readyQueue backed by the given store.
func newReadyQueue(s *store, maxTopics int) *readyQueue {
	return &readyQueue{
		callbacks: make(map[string]callbackEntry),
		maxTopics: maxTopics,
		store:     s,
	}
}

// registerCallback registers fn as the handler for the given topic and starts
// a drain goroutine that continuously pops tasks from the Redis ready list and
// invokes fn.
func (rq *readyQueue) registerCallback(topic string, fn func(context.Context, *Task) error) error {
	rq.mu.Lock()
	defer rq.mu.Unlock()

	if _, exists := rq.callbacks[topic]; exists {
		return ErrTopicConflict
	}
	if len(rq.callbacks) >= rq.maxTopics {
		return ErrTooManyTopics
	}

	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})

	rq.callbacks[topic] = callbackEntry{
		fn:     fn,
		cancel: cancel,
		done:   done,
	}

	go rq.drain(ctx, done, topic, fn)
	return nil
}

// drainPopTimeout is the BLPOP block duration used inside drain goroutines.
// go-redis enforces a minimum of 1s. Using a short timeout lets the goroutine
// check context cancellation frequently.
const drainPopTimeout = 1 * time.Second

// drain is the per-topic goroutine. It repeatedly calls PopTask (BLPOP) to
// receive tasks in StateActive, invokes fn, and on success auto-finishes the
// task. If fn returns an error the task is left in StateActive so that the TTR
// timer can trigger redelivery.
func (rq *readyQueue) drain(ctx context.Context, done chan struct{}, topic string, fn func(context.Context, *Task) error) {
	defer close(done)
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		task, err := rq.store.PopTask(ctx, topic, drainPopTimeout)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			continue
		}
		if task == nil {
			// BLPOP timeout, loop back and check ctx
			continue
		}

		if fnErr := fn(ctx, task); fnErr == nil {
			_ = rq.store.FinishTask(ctx, task.Topic, task.ID)
		}
	}
}

// hasCallback returns true when a callback is registered for topic.
func (rq *readyQueue) hasCallback(topic string) bool {
	rq.mu.RLock()
	defer rq.mu.RUnlock()
	_, ok := rq.callbacks[topic]
	return ok
}

// stopAll cancels every drain goroutine and waits for them to exit.
func (rq *readyQueue) stopAll() {
	rq.mu.Lock()
	entries := make([]callbackEntry, 0, len(rq.callbacks))
	for _, entry := range rq.callbacks {
		entry.cancel()
		entries = append(entries, entry)
	}
	rq.mu.Unlock()

	// Wait for all drain goroutines to exit (bounded by BLPOP timeout = 1s)
	for _, entry := range entries {
		<-entry.done
	}
}
