package token_bucket

import (
	"sync"
	"time"
)

var _ TokenBucket = (*TokenBucketLazy)(nil)

// ---- lazy refill under one mutex ----

type TokenBucketLazy struct {
	bucket, bucketSize, refillCount int
	mu                              sync.Mutex
	refillInterval                  time.Duration
	lastRefill                      time.Time
}

func NewTokenBucketLazy(bucketSize, refillCount int, refillInterval time.Duration) *TokenBucketLazy {
	tb := &TokenBucketLazy{
		bucket:         bucketSize,
		bucketSize:     bucketSize,
		refillCount:    refillCount,
		refillInterval: refillInterval,
		mu:             sync.Mutex{},
		lastRefill:     time.Now(),
	}

	return tb
}

func (tb *TokenBucketLazy) Allow() bool {
	now := time.Now() // hoisted OUT of the critical section
	tb.mu.Lock()
	defer tb.mu.Unlock()
	if n := int(now.Sub(tb.lastRefill) / tb.refillInterval); n > 0 {
		tb.bucket = min(tb.bucketSize, tb.bucket+n*tb.refillCount)
		// Adding to lastRefill instead of directly assigning `time.Now()`:
		// 	integer division throws away the remainder, and `= now` would throw away the tokens that remainder represents.
		//	detailed explanation in docs/lib/rate_limiter/token_bucket/token_bucket_lazy.md
		tb.lastRefill = tb.lastRefill.Add(time.Duration(n) * tb.refillInterval)
	}

	if tb.bucket > 0 {
		tb.bucket--
		return true
	}
	return false
}
