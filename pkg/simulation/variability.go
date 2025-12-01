package simulation

import (
	"math/rand"
	"time"
)

type LatencyVarier interface {
	AddVariability(base time.Duration, p99 time.Duration) time.Duration
}

var _ LatencyVarier = (*RandomLatencyVarier)(nil)

type RandomLatencyVarier struct {
	rand      *rand.Rand
	maxJitter time.Duration
}

func NewRandomLatencyVarier(rand rand.Rand, maxJitter time.Duration) *RandomLatencyVarier {
	return &RandomLatencyVarier{
		rand:      &rand,
		maxJitter: maxJitter,
	}
}

func (lv *RandomLatencyVarier) AddVariability(base time.Duration, p99 time.Duration) time.Duration {
	float := lv.rand.Float64()

	if float < 0.001 { // 0.1% of the time
		return 10 * p99
	}
	if float < 0.01 { // remaining 0.9% to reach 1%
		return p99
	}

	// else: 99% of the time - base with random jitter
	return max(0, jitter(lv.rand, base, lv.maxJitter)) // clamp to 0 to prevent negative durations when maxJitter > base
}

// jitter adds random jitter to base, returning base +/- maxJitter.
func jitter(r *rand.Rand, base time.Duration, maxJitter time.Duration) time.Duration {
	// Int63n(int64(2*maxJitter)) → random value in [0, 2*maxJitter)
	// subtract maxJitter → shifts range to [-maxJitter, +maxJitter)
	jitter := time.Duration(r.Int63n(int64(2*maxJitter))) - maxJitter
	return base + jitter
}
