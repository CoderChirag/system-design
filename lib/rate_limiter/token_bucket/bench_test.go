package token_bucket

import (
	"testing"
	"time"
)

/*
Ticker & Lazy Benchmark Execution: go test -bench='Ticker|Lazy' -benchtime=2s -cpu=8 ./...
*/

/*
Benchmark Ticker Parallel Testing bucket size: 1 Billion, refill rate: 1 Million / 1ms
*/
func BenchmarkTickerParallel(b *testing.B) {
	tb := NewTokenBucketTicker(1_000_000_000, 1_000_000, time.Millisecond)
	defer tb.Stop()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tb.Allow()
		}
	})
}

/*
Benchmark Lazy Parallel Testing bucket size: 1 Billion, refill rate: 1 Million / 1ms
*/
func BenchmarkLazyParallel(b *testing.B) {
	tb := NewTokenBucketLazy(1_000_000_000, 1_000_000, time.Millisecond)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			tb.Allow()
		}
	})
}

/*
Benchmark Lazy Serial Testing bucket size: 1 Billion, refill rate: 1 Million / 1ms
*/
func BenchmarkLazySerial(b *testing.B) {
	tb := NewTokenBucketLazy(1_000_000_000, 1_000_000, time.Millisecond)
	for b.Loop() {
		tb.Allow()
	}
}
