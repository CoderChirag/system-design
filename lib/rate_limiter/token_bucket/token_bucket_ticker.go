package token_bucket

import (
	"sync"
	"time"
)

var _ TokenBucket = (*TokenBucketTicker)(nil)

// ---- ticker-based token bucket algorithm ----

type TokenBucketTicker struct {
	bucket, bucketSize, refillCount int
	mu                              sync.Mutex
	stop                            chan bool
}

func NewTokenBucketTicker(bucketSize, refillCount int, refillInterval time.Duration) *TokenBucketTicker {
	tb := &TokenBucketTicker{
		bucket:      bucketSize,
		bucketSize:  bucketSize,
		refillCount: refillCount,
		mu:          sync.Mutex{},
		stop:        make(chan bool),
	}
	tb.startRefiller(refillInterval)
	return tb
}

func (tb *TokenBucketTicker) startRefiller(interval time.Duration) {
	go func() {
		tk := time.NewTicker(interval)
		defer tk.Stop()
		for {
			select {
			case <-tk.C:
				tb.mu.Lock()
				tb.bucket = min(tb.bucketSize, tb.bucket+tb.refillCount)
				tb.mu.Unlock()
			case <-tb.stop:
				return
			}
		}
	}()
}

func (tb *TokenBucketTicker) Stop() {
	close(tb.stop)
}

func (tb *TokenBucketTicker) Allow() bool {
	tb.mu.Lock()
	defer tb.mu.Unlock()
	if tb.bucket > 0 {
		tb.bucket--
		return true
	}
	return false
}
