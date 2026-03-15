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
//
// Returns ErrTopicConflict if a callback is already registered for the topic,
// or ErrTooManyTopics if the topic cap has been reached.
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

	rq.callbacks[topic] = callbackEntry{
		fn:     fn,
		cancel: cancel,
	}

	go rq.drain(ctx, topic, fn)
	return nil
}

// drainPopTimeout is the BLPOP block duration used inside drain goroutines.
// Using a finite timeout lets the goroutine observe context cancellation
// without blocking forever on an empty list.
const drainPopTimeout = 2 * time.Second

// drain is the per-topic goroutine. It repeatedly calls PopTask (BLPOP) to
// receive tasks in StateActive, invokes fn, and on success auto-finishes the
// task. If fn returns an error the task is left in StateActive so that the TTR
// timer can trigger redelivery.
func (rq *readyQueue) drain(ctx context.Context, topic string, fn func(context.Context, *Task) error) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
		}

		// PopTask sets the task state to StateActive before returning.
		task, err := rq.store.PopTask(ctx, topic, drainPopTimeout)
		if err != nil {
			// Context cancelled or Redis error — exit the goroutine.
			if ctx.Err() != nil {
				return
			}
			// Transient Redis error; continue and retry.
			continue
		}
		if task == nil {
			// Timeout elapsed with no item; loop back to BLPOP.
			continue
		}

		// Invoke the user callback.
		if fnErr := fn(ctx, task); fnErr == nil {
			// Success: mark the task finished. Ignore the error here as the
			// task will naturally expire anyway; best-effort finish is fine.
			_ = rq.store.FinishTask(ctx, task.Topic, task.ID)
		}
		// On callback error: do nothing. The TTR wheel entry will fire
		// ReadyTask again for redelivery.
	}
}

// hasCallback returns true when a callback is registered for topic.
func (rq *readyQueue) hasCallback(topic string) bool {
	rq.mu.RLock()
	defer rq.mu.RUnlock()
	_, ok := rq.callbacks[topic]
	return ok
}

// stopAll cancels every drain goroutine.
func (rq *readyQueue) stopAll() {
	rq.mu.Lock()
	defer rq.mu.Unlock()

	for _, entry := range rq.callbacks {
		entry.cancel()
	}
}
