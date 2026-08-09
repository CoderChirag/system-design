package main

import (
	"os"
	rate_limiter_client "system_design/client/rate_limiter"
	"system_design/lib/rate_limiter/fixed_window"
	"system_design/lib/rate_limiter/leaking_bucket"
	"system_design/lib/rate_limiter/sliding_window"
	"system_design/lib/rate_limiter/token_bucket"
	"time"
)

func main() {
	rateLimiterType := os.Args[1]
	switch rateLimiterType {
	case "-token_bucket_ticker":
		tb := token_bucket.NewTokenBucketTicker(5, 2, time.Second)
		defer tb.Stop()
		rate_limiter_client.TokenBucketBurst(tb)
	case "-token_bucket_lazy":
		tb := token_bucket.NewTokenBucketLazy(5, 2, time.Second)
		rate_limiter_client.TokenBucketBurst(tb)
	case "-leaking_bucket":
		lb := leaking_bucket.NewLeakingBucket(5, 2, time.Second)
		defer lb.StopAndWait()
		rate_limiter_client.LeakingBucketburst(lb)
	case "-fixed_window_counter":
		fwc := fixed_window.NewFixedWindowCounter(5, time.Second)
		rate_limiter_client.FixedWindowCounterBurst(fwc)
	case "-sliding_window_log":
		sw := sliding_window.NewSlidingWindow(5, time.Second)
		rate_limiter_client.SlidingWindowLogBurst(sw)
	}

}
