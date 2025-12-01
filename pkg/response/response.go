package response

import "time"

type Response struct {
	RequestID string
	ServerID  string
	Success   bool
	Err       error

	// Timing breakdown
	Latency        time.Duration // total time (queue + processing)
	QueueWait      time.Duration // time spent in queue
	ProcessingTime time.Duration // actual processing time
}
