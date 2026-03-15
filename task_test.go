package seqdelay

import (
	"testing"
	"time"
)

func TestTask_Validate(t *testing.T) {
	tests := []struct {
		name    string
		task    Task
		wantErr bool
	}{
		{"valid", Task{ID: "1", Topic: "t", Delay: time.Second}, false},
		{"missing id", Task{Topic: "t", Delay: time.Second}, true},
		{"missing topic", Task{ID: "1", Delay: time.Second}, true},
		{"zero delay", Task{ID: "1", Topic: "t"}, true},
		{"negative delay", Task{ID: "1", Topic: "t", Delay: -1}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.task.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestTask_DefaultTTR(t *testing.T) {
	task := Task{ID: "1", Topic: "t", Delay: time.Second}
	task.Validate()
	if task.TTR != 30*time.Second {
		t.Errorf("default TTR = %v, want 30s", task.TTR)
	}
}

func TestTask_RedisKeys(t *testing.T) {
	if got := taskKey("orders", "o1"); got != "seqdelay:{orders}:task:o1" {
		t.Errorf("taskKey = %s", got)
	}
	if got := readyKey("orders"); got != "seqdelay:{orders}:ready" {
		t.Errorf("readyKey = %s", got)
	}
	if got := indexKey("orders"); got != "seqdelay:{orders}:index" {
		t.Errorf("indexKey = %s", got)
	}
}

func TestTaskState_String(t *testing.T) {
	cases := map[TaskState]string{
		StateDelayed:   "delayed",
		StateReady:     "ready",
		StateActive:    "active",
		StateFinished:  "finished",
		StateCancelled: "cancelled",
		TaskState(99):  "unknown",
	}
	for state, want := range cases {
		if got := state.String(); got != want {
			t.Errorf("%d.String() = %s, want %s", state, got, want)
		}
	}
}
