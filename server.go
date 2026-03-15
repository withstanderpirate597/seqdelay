package seqdelay

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"
)

// ---------------------------------------------------------------------------
// Server
// ---------------------------------------------------------------------------

// Server exposes the Queue over HTTP using a plain net/http ServeMux.
// All request/response bodies are JSON.
type Server struct {
	queue  *Queue
	addr   string
	server *http.Server
}

// ServerOption configures a Server.
type ServerOption func(*Server)

// WithServerAddr sets the TCP listen address (default ":8080").
func WithServerAddr(addr string) ServerOption {
	return func(s *Server) { s.addr = addr }
}

// NewServer creates a Server that wraps the given Queue.
func NewServer(q *Queue, opts ...ServerOption) *Server {
	s := &Server{
		queue: q,
		addr:  ":8080",
	}
	for _, o := range opts {
		o(s)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /add", s.handleAdd)
	mux.HandleFunc("POST /pop", s.handlePop)
	mux.HandleFunc("POST /finish", s.handleFinish)
	mux.HandleFunc("POST /cancel", s.handleCancel)
	mux.HandleFunc("GET /get", s.handleGet)
	mux.HandleFunc("GET /stats", s.handleStats)

	s.server = &http.Server{
		Addr:    s.addr,
		Handler: mux,
	}
	return s
}

// ListenAndServe starts the HTTP server. It blocks until the server stops.
func (s *Server) ListenAndServe() error {
	return s.server.ListenAndServe()
}

// Shutdown gracefully shuts down the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

// ---------------------------------------------------------------------------
// Request / response types
// ---------------------------------------------------------------------------

type apiResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message,omitempty"`
	Data    any    `json:"data,omitempty"`
}

type addRequest struct {
	Topic      string `json:"topic"`
	ID         string `json:"id"`
	Body       string `json:"body"`
	DelayMS    int64  `json:"delay_ms"`
	TTRMS      int64  `json:"ttr_ms"`
	MaxRetries int    `json:"max_retries"`
}

type popRequest struct {
	Topic     string `json:"topic"`
	TimeoutMS int64  `json:"timeout_ms"`
}

type topicIDRequest struct {
	Topic string `json:"topic"`
	ID    string `json:"id"`
}

type taskData struct {
	ID    string `json:"id"`
	Topic string `json:"topic"`
	Body  string `json:"body"`
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// writeJSON encodes v as JSON with the given HTTP status code.
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

// ok writes a standard success response.
func ok(w http.ResponseWriter) {
	writeJSON(w, http.StatusOK, apiResponse{Code: 0, Message: "ok"})
}

// errResponse maps a domain error to an HTTP status code and writes the
// corresponding JSON error response.
func errResponse(w http.ResponseWriter, err error) {
	var status int
	switch {
	case errors.Is(err, ErrDuplicateTask):
		status = http.StatusConflict
	case errors.Is(err, ErrTaskNotFound):
		status = http.StatusNotFound
	case errors.Is(err, ErrInvalidTask), errors.Is(err, ErrInvalidDelay):
		status = http.StatusBadRequest
	case errors.Is(err, ErrTopicConflict):
		status = http.StatusConflict
	case errors.Is(err, ErrClosed):
		status = http.StatusServiceUnavailable
	default:
		status = http.StatusInternalServerError
	}
	writeJSON(w, status, apiResponse{Code: status, Message: err.Error()})
}

// decodeJSON decodes the request body into v. Returns false and writes a 400
// response on failure.
func decodeJSON(w http.ResponseWriter, r *http.Request, v any) bool {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		writeJSON(w, http.StatusBadRequest, apiResponse{
			Code:    http.StatusBadRequest,
			Message: "invalid JSON: " + err.Error(),
		})
		return false
	}
	return true
}

// ---------------------------------------------------------------------------
// Handlers
// ---------------------------------------------------------------------------

// handleAdd handles POST /add.
// Body: {"topic":"x","id":"y","body":"...","delay_ms":1000,"ttr_ms":30000,"max_retries":3}
func (s *Server) handleAdd(w http.ResponseWriter, r *http.Request) {
	var req addRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	task := &Task{
		ID:         req.ID,
		Topic:      req.Topic,
		Body:       []byte(req.Body),
		Delay:      time.Duration(req.DelayMS) * time.Millisecond,
		TTR:        time.Duration(req.TTRMS) * time.Millisecond,
		MaxRetries: req.MaxRetries,
	}

	if err := s.queue.Add(r.Context(), task); err != nil {
		errResponse(w, err)
		return
	}
	ok(w)
}

// handlePop handles POST /pop.
// Body: {"topic":"x","timeout_ms":30000}
func (s *Server) handlePop(w http.ResponseWriter, r *http.Request) {
	var req popRequest
	if !decodeJSON(w, r, &req) {
		return
	}

	// Override the queue's default pop timeout if the caller supplied one.
	ctx := r.Context()
	if req.TimeoutMS > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(req.TimeoutMS)*time.Millisecond)
		defer cancel()
	}

	task, err := s.queue.Pop(ctx, req.Topic)
	if err != nil {
		errResponse(w, err)
		return
	}
	if task == nil {
		// Timeout with no task.
		writeJSON(w, http.StatusOK, apiResponse{Code: 0, Data: nil})
		return
	}

	writeJSON(w, http.StatusOK, apiResponse{
		Code: 0,
		Data: taskData{
			ID:    task.ID,
			Topic: task.Topic,
			Body:  string(task.Body),
		},
	})
}

// handleFinish handles POST /finish.
// Body: {"topic":"x","id":"y"}
func (s *Server) handleFinish(w http.ResponseWriter, r *http.Request) {
	var req topicIDRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if err := s.queue.Finish(r.Context(), req.Topic, req.ID); err != nil {
		errResponse(w, err)
		return
	}
	ok(w)
}

// handleCancel handles POST /cancel.
// Body: {"topic":"x","id":"y"}
func (s *Server) handleCancel(w http.ResponseWriter, r *http.Request) {
	var req topicIDRequest
	if !decodeJSON(w, r, &req) {
		return
	}
	if err := s.queue.Cancel(r.Context(), req.Topic, req.ID); err != nil {
		errResponse(w, err)
		return
	}
	ok(w)
}

// handleGet handles GET /get?topic=x&id=y.
func (s *Server) handleGet(w http.ResponseWriter, r *http.Request) {
	topic := r.URL.Query().Get("topic")
	id := r.URL.Query().Get("id")

	if topic == "" || id == "" {
		writeJSON(w, http.StatusBadRequest, apiResponse{
			Code:    http.StatusBadRequest,
			Message: "topic and id query parameters are required",
		})
		return
	}

	task, err := s.queue.Get(r.Context(), topic, id)
	if err != nil {
		errResponse(w, err)
		return
	}

	writeJSON(w, http.StatusOK, apiResponse{Code: 0, Data: task})
}

// topicStats holds per-topic ready and active counts.
type topicStats struct {
	Ready  int64 `json:"ready"`
	Active int64 `json:"active"`
}

// handleStats handles GET /stats.
// Returns per-topic ready-list length and active task count derived from the
// Redis index.
func (s *Server) handleStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	topics, err := s.queue.store.ListTopics(ctx)
	if err != nil {
		errResponse(w, err)
		return
	}

	stats := make(map[string]topicStats, len(topics))
	for _, topic := range topics {
		ready, _ := s.queue.cfg.redisClient.LLen(ctx, readyKey(topic)).Result()

		tasks, _ := s.queue.store.LoadTopicTasks(ctx, topic)
		var active int64
		for _, t := range tasks {
			if t.State == StateActive {
				active++
			}
		}

		stats[topic] = topicStats{Ready: ready, Active: active}
	}

	writeJSON(w, http.StatusOK, apiResponse{
		Code: 0,
		Data: map[string]any{"topics": stats},
	})
}
