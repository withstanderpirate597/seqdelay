package seqdelay

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

// newTestQueue creates a Queue backed by the test Redis instance and starts it.
// The caller is responsible for calling q.Shutdown after the test.
func newTestQueue(t *testing.T) *Queue {
	t.Helper()
	client := getTestRedis(t)
	t.Cleanup(func() { client.Close() })

	q, err := New(
		WithRedisClient(client),
		WithTickInterval(5*time.Millisecond),
		WithPopTimeout(3*time.Second),
	)
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if err := q.Start(context.Background()); err != nil {
		t.Fatalf("Start: %v", err)
	}
	t.Cleanup(func() { q.Shutdown(context.Background()) }) //nolint:errcheck
	return q
}

// newTestServer returns an httptest.Server wrapping a Queue and registers
// cleanup.
func newTestServer(t *testing.T) (*httptest.Server, *Queue) {
	t.Helper()
	q := newTestQueue(t)
	srv := NewServer(q)
	ts := httptest.NewServer(srv.server.Handler)
	t.Cleanup(ts.Close)
	return ts, q
}

// postJSON sends a POST request with a JSON body and returns the response.
func postJSON(t *testing.T, ts *httptest.Server, path string, body any) *http.Response {
	t.Helper()
	data, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	resp, err := http.Post(ts.URL+path, "application/json", bytes.NewReader(data))
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	return resp
}

// decodeResp decodes the JSON response body into apiResponse.
func decodeResp(t *testing.T, resp *http.Response) apiResponse {
	t.Helper()
	defer resp.Body.Close()
	var r apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	return r
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// TestServer_Add returns HTTP 200 on a valid add request.
func TestServer_Add(t *testing.T) {
	ts, _ := newTestServer(t)

	resp := postJSON(t, ts, "/add", map[string]any{
		"topic":    "srv-add",
		"id":       "add-1",
		"body":     "hello",
		"delay_ms": 5000,
		"ttr_ms":   30000,
	})

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	r := decodeResp(t, resp)
	if r.Code != 0 {
		t.Errorf("expected code 0, got %d", r.Code)
	}
}

// TestServer_Add_Duplicate returns HTTP 409 when the same task is added twice.
func TestServer_Add_Duplicate(t *testing.T) {
	ts, _ := newTestServer(t)

	body := map[string]any{
		"topic":    "srv-dup",
		"id":       "dup-srv-1",
		"body":     "x",
		"delay_ms": 5000,
		"ttr_ms":   30000,
	}

	resp1 := postJSON(t, ts, "/add", body)
	resp1.Body.Close()
	if resp1.StatusCode != http.StatusOK {
		t.Fatalf("first add expected 200, got %d", resp1.StatusCode)
	}

	resp2 := postJSON(t, ts, "/add", body)
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusConflict {
		t.Errorf("duplicate add expected 409, got %d", resp2.StatusCode)
	}
}

// TestServer_Pop returns the popped task.
func TestServer_Pop(t *testing.T) {
	ts, q := newTestServer(t)
	ctx := context.Background()

	// Acquire leader lock so the timing wheel fires ReadyTask.
	q.lock.TryAcquire(ctx)

	// Add a task with a very short delay.
	task := &Task{
		ID:    "pop-srv-1",
		Topic: "srv-pop",
		Body:  []byte("pop-body"),
		Delay: 50 * time.Millisecond,
		TTR:   30 * time.Second,
	}
	if err := q.Add(ctx, task); err != nil {
		t.Fatalf("Add: %v", err)
	}

	// Wait for the timing wheel to move it to ready.
	time.Sleep(200 * time.Millisecond)

	resp := postJSON(t, ts, "/pop", map[string]any{
		"topic":      "srv-pop",
		"timeout_ms": 3000,
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var r struct {
		Code int      `json:"code"`
		Data taskData `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		t.Fatalf("decode pop response: %v", err)
	}
	if r.Data.ID != task.ID {
		t.Errorf("expected task ID %s, got %s", task.ID, r.Data.ID)
	}
}

// TestServer_Finish returns HTTP 200 after finishing an active task.
func TestServer_Finish(t *testing.T) {
	ts, q := newTestServer(t)
	ctx := context.Background()

	q.lock.TryAcquire(ctx)

	task := &Task{
		ID:    "finish-srv-1",
		Topic: "srv-finish",
		Body:  []byte("done"),
		Delay: 50 * time.Millisecond,
		TTR:   30 * time.Second,
	}
	if err := q.Add(ctx, task); err != nil {
		t.Fatalf("Add: %v", err)
	}

	time.Sleep(200 * time.Millisecond)

	// Pop via the store directly to avoid an extra HTTP round trip.
	_, err := q.Pop(ctx, task.Topic)
	if err != nil {
		t.Fatalf("Pop: %v", err)
	}

	resp := postJSON(t, ts, "/finish", map[string]any{
		"topic": "srv-finish",
		"id":    "finish-srv-1",
	})
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
}

// TestServer_Get returns the full task JSON.
func TestServer_Get(t *testing.T) {
	ts, q := newTestServer(t)
	ctx := context.Background()

	task := &Task{
		ID:    "get-srv-1",
		Topic: "srv-get",
		Body:  []byte("look at me"),
		Delay: 5 * time.Second,
		TTR:   30 * time.Second,
	}
	if err := q.Add(ctx, task); err != nil {
		t.Fatalf("Add: %v", err)
	}

	url := fmt.Sprintf("%s/get?topic=srv-get&id=get-srv-1", ts.URL)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("GET /get: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var r struct {
		Code int   `json:"code"`
		Data *Task `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		t.Fatalf("decode get response: %v", err)
	}
	if r.Data == nil {
		t.Fatal("data field is nil")
	}
	if r.Data.ID != task.ID {
		t.Errorf("expected ID %s, got %s", task.ID, r.Data.ID)
	}
}

// TestServer_Stats returns stats JSON.
func TestServer_Stats(t *testing.T) {
	ts, q := newTestServer(t)
	ctx := context.Background()

	// Seed a task so the topic appears in stats.
	task := &Task{
		ID:    "stats-1",
		Topic: "srv-stats",
		Body:  []byte("stat-me"),
		Delay: 10 * time.Second,
		TTR:   30 * time.Second,
	}
	if err := q.Add(ctx, task); err != nil {
		t.Fatalf("Add: %v", err)
	}

	resp, err := http.Get(ts.URL + "/stats")
	if err != nil {
		t.Fatalf("GET /stats: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}

	var r apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&r); err != nil {
		t.Fatalf("decode stats response: %v", err)
	}
	if r.Code != 0 {
		t.Errorf("expected code 0, got %d", r.Code)
	}
	if r.Data == nil {
		t.Error("data field is nil")
	}
}

// TestServer_InvalidJSON returns HTTP 400 on malformed request body.
func TestServer_InvalidJSON(t *testing.T) {
	ts, _ := newTestServer(t)

	resp, err := http.Post(ts.URL+"/add", "application/json",
		strings.NewReader("{invalid json"))
	if err != nil {
		t.Fatalf("POST /add: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}
