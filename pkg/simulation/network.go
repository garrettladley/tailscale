package simulation

import (
	"math/rand"
	"time"
)

type NetworkDelayer interface {
	NetworkDelay() time.Duration
}

var _ NetworkDelayer = (*RandomNetworkDelayer)(nil)

type RandomNetworkDelayer struct {
	rand             *rand.Rand
	baseNetworkDelay time.Duration
	maxJitter        time.Duration
}

func NewRandomNetworkDelayer(rand *rand.Rand, baseNetworkDelay time.Duration, maxJitter time.Duration) *RandomNetworkDelayer {
	return &RandomNetworkDelayer{
		rand:      rand,
		maxJitter: maxJitter,
	}
}

func (rn *RandomNetworkDelayer) NetworkDelay() time.Duration {
	return rn.baseNetworkDelay + time.Duration(rn.rand.Int63n(int64(rn.maxJitter)))
}
