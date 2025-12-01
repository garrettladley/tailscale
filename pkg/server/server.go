package server

import (
	"context"
	"errors"
	"math/rand"
	"sync"
	"time"

	"github.com/garrettladley/tailscale/pkg/request"
	"github.com/garrettladley/tailscale/pkg/response"
	"github.com/garrettladley/tailscale/pkg/simulation"
)

var ErrCancelled = errors.New("request cancelled")

type Config struct {
	ID           string
	BaseLatency  time.Duration
	P99Latency   time.Duration
	Rand         *rand.Rand
	FailureRate  float64
	QueueSize    uint
	GetRequestID func(request.Request) string
}

type Server struct {
	cfg           Config
	latencyVarier simulation.LatencyVarier

	queue chan *workItem

	mu      sync.Mutex
	pending map[string]*workItem // requestID -> workItem (if still queued)

	done chan struct{}
}

type workItem struct {
	req      request.Request
	ctx      context.Context
	respChan chan response.Response
	enqueued time.Time
}

func NewServer(cfg Config, lv simulation.LatencyVarier) *Server {
	s := &Server{
		cfg:           cfg,
		latencyVarier: lv,
		queue:         make(chan *workItem, cfg.QueueSize),
		pending:       make(map[string]*workItem),
		done:          make(chan struct{}),
	}
	go s.worker()
	return s
}

// Handle enqueues a request and blocks until processed or context cancelled.
func (s *Server) Handle(ctx context.Context, req request.Request) response.Response {
	item := &workItem{
		req:      req,
		ctx:      ctx,
		respChan: make(chan response.Response, 1), // buffered so worker doesn't block
		enqueued: time.Now(),
	}

	reqID := s.cfg.GetRequestID(req)

	// track in pending map for cancellation support
	s.mu.Lock()
	s.pending[reqID] = item
	s.mu.Unlock()

	// try enqueue
	select {
	case s.queue <- item:
	case <-ctx.Done():
		// context cancelled before we could enqueue
		s.mu.Lock()
		delete(s.pending, reqID)
		s.mu.Unlock()
		return response.Response{
			RequestID: reqID,
			ServerID:  s.cfg.ID,
			Success:   false,
			Err:       ctx.Err(),
		}
	}

	// wait for response or cancellation
	select {
	case resp := <-item.respChan:
		return resp
	case <-ctx.Done():
		// context cancelled while waiting
		// note: request might still be processed, but caller doesn't care
		return response.Response{
			RequestID: reqID,
			ServerID:  s.cfg.ID,
			Success:   false,
			Err:       ctx.Err(),
		}
	}
}

// Cancel attempts to cancel a queued request (for tied requests).
// Returns true if the request was still queued and got cancelled.
func (s *Server) Cancel(reqID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, exists := s.pending[reqID]
	if !exists {
		return false // already processed or never existed
	}

	// mark as cancelled by sending a cancelled response
	// only works if still in queue (not yet picked up by worker)
	select {
	case item.respChan <- response.Response{
		RequestID: reqID,
		ServerID:  s.cfg.ID,
		Success:   false,
		Err:       ErrCancelled,
	}:
		delete(s.pending, reqID)
		return true
	default:
		// response channel already has something - worker got there first
		return false
	}
}

// worker processes requests from the queue
func (s *Server) worker() {
	for {
		select {
		case item := <-s.queue:
			s.processItem(item)
		case <-s.done:
			return
		}
	}
}

func (s *Server) processItem(item *workItem) {
	reqID := s.cfg.GetRequestID(item.req)

	// remove from pending map - we're now processing
	s.mu.Lock()
	_, stillPending := s.pending[reqID]
	delete(s.pending, reqID)
	s.mu.Unlock()

	// if it was cancelled while queued, the Cancel() method already sent a response
	if !stillPending {
		return
	}

	// check if context was cancelled while waiting in queue
	if item.ctx.Err() != nil {
		item.respChan <- response.Response{
			RequestID: reqID,
			ServerID:  s.cfg.ID,
			Success:   false,
			Err:       item.ctx.Err(),
		}
		return
	}

	queueWait := time.Since(item.enqueued)

	processingTime := s.latencyVarier.AddVariability(s.cfg.BaseLatency, s.cfg.P99Latency)

	timer := time.NewTimer(processingTime)
	defer timer.Stop()

	select {
	case <-timer.C:
	case <-item.ctx.Done():
		item.respChan <- response.Response{
			RequestID: reqID,
			ServerID:  s.cfg.ID,
			Success:   false,
			Err:       item.ctx.Err(),
		}
		return
	}

	isSuccess := s.cfg.Rand.Float64() >= s.cfg.FailureRate

	item.respChan <- response.Response{
		RequestID:      reqID,
		ServerID:       s.cfg.ID,
		Success:        isSuccess,
		Latency:        queueWait + processingTime,
		QueueWait:      queueWait,
		ProcessingTime: processingTime,
	}
}

// Shutdown gracefully stops the server
func (s *Server) Shutdown() {
	close(s.done)
}

// QueueDepth returns current number of items waiting (useful for load balancing)
func (s *Server) QueueDepth() int {
	return len(s.queue)
}
