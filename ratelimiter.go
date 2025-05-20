package tgnotificator

import (
	"sync"
	"time"
)

type rateLimiter struct {
	rps        int
	interval   time.Duration
	requests   int
	retryAfter time.Time
	mu         sync.Mutex
}

func newRateLimiter(rps int, interval time.Duration) *rateLimiter {
	return &rateLimiter{
		rps:        rps,
		interval:   interval,
		requests:   0,
		retryAfter: time.Now().Add(interval),
	}
}

func (rl *rateLimiter) allowRequest() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	if time.Now().After(rl.retryAfter) {
		rl.requests = 0
		rl.retryAfter = time.Now().Add(rl.interval)
	}

	if rl.requests < rl.rps {
		rl.requests++
		return true
	}

	return false
}

func (rl *rateLimiter) currentState() (int, time.Duration) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	return rl.requests, time.Until(rl.retryAfter)
}
