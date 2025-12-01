package core

import "time"

type Request struct {
	ID        string    `json:"id"`
	Timestamp time.Time `json:"ts"`
	Payload   any       `json:"payload"`
}

type Response struct {
	RequestID string        `json:"request_id"`
	Latency   time.Duration `json:"latency"`
	ServerID  string        `json:"server_id"`
	Success   bool          `json:"success"`
}

type SevrverConfig struct {
	BaseLatency time.Duration `json:"base_latency"`
	P99Latency  time.Duration `json:"p99_latency"`
	// 0.01 -> 1%
	FailureRate float64 `json:"failure_pct"`
}
