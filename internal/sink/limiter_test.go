package sink

import (
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	rate := 100 // 100 bytes/sec
	rl := NewRateLimiter(rate)

	// Consuming 50 bytes should be allowed
	if !rl.Allow(50) {
		t.Error("Expected to allow 50 bytes from fresh bucket")
	}

	// Consuming another 40 bytes should be allowed
	if !rl.Allow(40) {
		t.Error("Expected to allow 40 bytes")
	}

	// Consuming 20 bytes should be rejected (bucket has ~10 left)
	if rl.Allow(20) {
		t.Error("Expected to reject 20 bytes (bucket nearly empty)")
	}

	// Wait for refill (0.5s -> 50 bytes)
	time.Sleep(500 * time.Millisecond)

	// Now 20 bytes should be allowed
	if !rl.Allow(20) {
		t.Error("Expected to allow 20 bytes after refill")
	}
}

func TestRateLimiter_Unlimited(t *testing.T) {
	rl := NewRateLimiter(0)
	if !rl.Allow(1000000) {
		t.Error("Expected to allow large amount when unlimited")
	}
}
