package sink

import (
	"sync"
	"time"
)

// RateLimiter implements a token bucket rate limiter.
type RateLimiter struct {
	rate       int
	bucket     float64
	lastRefill time.Time
	mu         sync.Mutex
}

// NewRateLimiter creates a new RateLimiter with the specified rate (bytes/second).
func NewRateLimiter(rate int) *RateLimiter {
	return &RateLimiter{
		rate:       rate,
		bucket:     float64(rate),
		lastRefill: time.Now(),
	}
}

// Allow checks if the request of size n bytes is allowed.
func (rl *RateLimiter) Allow(n int) bool {
	if rl.rate == 0 {
		return true
	}

	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(rl.lastRefill).Seconds()
	rl.lastRefill = now

	rl.bucket += elapsed * float64(rl.rate)
	if rl.bucket > float64(rl.rate) {
		rl.bucket = float64(rl.rate)
	}

	if rl.bucket >= float64(n) {
		rl.bucket -= float64(n)
		return true
	}

	return false
}
