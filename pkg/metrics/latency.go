package metrics

import (
	"slices"
	"sync"
	"time"
)

type LatencyTracker struct {
	mu      sync.Mutex
	samples []time.Duration
	index   uint
	full    bool
	maxSize uint
}

type LatencySummary struct {
	P50   time.Duration `json:"p50"`
	P95   time.Duration `json:"p95"`
	P99   time.Duration `json:"p99"`
	P99_9 time.Duration `json:"p99.9"`
}

func NewLatencyTracker(maxSize uint) *LatencyTracker {
	return &LatencyTracker{
		samples: make([]time.Duration, maxSize),
		maxSize: maxSize,
	}
}

func (lt *LatencyTracker) Record(d time.Duration) {
	lt.mu.Lock()
	lt.samples[lt.index] = d
	lt.index++
	if lt.index >= lt.maxSize {
		lt.index = 0
		lt.full = true
	}
	lt.mu.Unlock()
}

func (lt *LatencyTracker) Percentile(p float64) time.Duration {
	if p < 0 || p > 1 {
		return 0
	}

	lt.mu.Lock()
	size := lt.index
	if lt.full {
		size = lt.maxSize
	}
	if size == 0 {
		lt.mu.Unlock()
		return 0
	}

	// copy samples while holding lock
	sorted := make([]time.Duration, size)
	copy(sorted, lt.samples[:size])
	lt.mu.Unlock()

	// sort outside the lock
	slices.Sort(sorted)

	return percentile(sorted, p)
}

func (lt *LatencyTracker) Summary() LatencySummary {
	lt.mu.Lock()
	size := lt.index
	if lt.full {
		size = lt.maxSize
	}
	if size == 0 {
		lt.mu.Unlock()
		return LatencySummary{}
	}

	sorted := make([]time.Duration, size)
	copy(sorted, lt.samples[:size])
	lt.mu.Unlock()

	slices.Sort(sorted)

	return LatencySummary{
		P50:   percentile(sorted, 0.50),
		P95:   percentile(sorted, 0.95),
		P99:   percentile(sorted, 0.99),
		P99_9: percentile(sorted, 0.999),
	}
}

// p=0.99 -> 99th percentile
func percentile(sorted []time.Duration, p float64) time.Duration {
	lenSorted := len(sorted)
	if lenSorted == 0 {
		return 0
	}
	if lenSorted == 1 {
		return sorted[0]
	}

	// find the position: p * (n-1)
	pos := p * float64(len(sorted)-1)
	lower := int(pos)
	upper := lower + 1

	if upper >= len(sorted) {
		return sorted[len(sorted)-1]
	}

	fraction := pos - float64(lower)
	interpolated := sorted[lower] + time.Duration(float64(sorted[upper]-sorted[lower])*fraction)

	return interpolated
}
