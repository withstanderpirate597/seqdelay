package seqdelay

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestReadyQueue_RegisterCallback(t *testing.T) {
	client := getTestRedis(t)
	s := newStore(client)
	rq := newReadyQueue(s, 1024)
	defer rq.stopAll()

	called := make(chan string, 1)
	err := rq.registerCallback("test-topic", func(ctx context.Context, task *Task) error {
		called <- task.ID
		return nil
	})
	if err != nil {
		t.Fatalf("registerCallback error: %v", err)
	}

	// Push a task to ready list and verify callback fires
	ctx := context.Background()
	task := &Task{
		ID: "t1", Topic: "test-topic", Body: []byte("{}"),
		Delay: time.Second, TTR: 30 * time.Second,
		State: StateDelayed, CreatedAt: time.Now(),
	}
	s.SaveTask(ctx, task)
	s.ReadyTask(ctx, "test-topic", "t1")

	select {
	case id := <-called:
		if id != "t1" {
			t.Errorf("callback got id=%s, want t1", id)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("timeout waiting for callback")
	}
}

func TestReadyQueue_TopicConflict(t *testing.T) {
	client := getTestRedis(t)
	s := newStore(client)
	rq := newReadyQueue(s, 1024)
	defer rq.stopAll()

	fn := func(ctx context.Context, task *Task) error { return nil }
	if err := rq.registerCallback("dup", fn); err != nil {
		t.Fatal(err)
	}
	if err := rq.registerCallback("dup", fn); err != ErrTopicConflict {
		t.Errorf("expected ErrTopicConflict, got %v", err)
	}
}

func TestReadyQueue_MaxTopics(t *testing.T) {
	client := getTestRedis(t)
	s := newStore(client)
	rq := newReadyQueue(s, 2) // max 2 topics
	defer rq.stopAll()

	fn := func(ctx context.Context, task *Task) error { return nil }
	rq.registerCallback("t1", fn)
	rq.registerCallback("t2", fn)
	if err := rq.registerCallback("t3", fn); err != ErrTooManyTopics {
		t.Errorf("expected ErrTooManyTopics, got %v", err)
	}
}

func TestReadyQueue_HasCallback(t *testing.T) {
	client := getTestRedis(t)
	s := newStore(client)
	rq := newReadyQueue(s, 1024)
	defer rq.stopAll()

	if rq.hasCallback("x") {
		t.Error("hasCallback should return false for unregistered topic")
	}
	rq.registerCallback("x", func(ctx context.Context, task *Task) error { return nil })
	if !rq.hasCallback("x") {
		t.Error("hasCallback should return true after register")
	}
}

func TestReadyQueue_CallbackError_NoAutoFinish(t *testing.T) {
	client := getTestRedis(t)
	s := newStore(client)
	rq := newReadyQueue(s, 1024)
	defer rq.stopAll()

	var calls atomic.Int32
	rq.registerCallback("err-topic", func(ctx context.Context, task *Task) error {
		calls.Add(1)
		return ErrInvalidState // simulate failure
	})

	ctx := context.Background()
	task := &Task{
		ID: "t1", Topic: "err-topic", Body: []byte("{}"),
		Delay: time.Second, TTR: 30 * time.Second,
		State: StateDelayed, CreatedAt: time.Now(),
	}
	s.SaveTask(ctx, task)
	s.ReadyTask(ctx, "err-topic", "t1")

	// Wait for callback to fire
	time.Sleep(3 * time.Second)

	// Task should NOT be finished (callback returned error)
	got, err := s.GetTask(ctx, "err-topic", "t1")
	if err != nil {
		t.Fatalf("GetTask error: %v", err)
	}
	// State should be ACTIVE (popped by drain goroutine) not FINISHED
	if got.State == StateFinished {
		t.Error("task should NOT be auto-finished when callback returns error")
	}
}
