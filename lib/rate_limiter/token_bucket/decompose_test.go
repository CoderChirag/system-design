package token_bucket

import (
	"sync"
	"testing"
	"time"
)

var tsink time.Time

/*
Execution: go test -bench='Empty|TimeNow|MutexOnly|MutexPlusMath' -benchtime=2s -cpu=8 ./...
*/

// ---- 1. Cost decomposition: peel the operation apart layer by layer ----

// Baseline: what does the benchmark harness itself cost per iteration?
func BenchmarkEmptyParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
		}
	})
}

// Just reading the clock. No sharing at all.
func BenchmarkTimeNowParallel(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		var t time.Time
		for pb.Next() {
			t = time.Now()
		}
		tsink = t
	})
}

// Just the lock handoff, contended by all threads. Empty critical section.
func BenchmarkMutexOnlyParallel(b *testing.B) {
	var mu sync.Mutex
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			mu.Lock()
			mu.Unlock()
		}
	})
}

func BenchmarkTimeNowSerial(b *testing.B) {
	var t time.Time
	for b.Loop() {
		t = time.Now()
	}
	tsink = t
}

func BenchmarkMutexOnlySerial(b *testing.B) {
	var mu sync.Mutex
	for b.Loop() {
		mu.Lock()
		mu.Unlock()
	}
}

// Lock + the integer arithmetic, but NO clock read. (~= the ticker design's
// request path.)
func BenchmarkMutexPlusMathParallel(b *testing.B) {
	l := NewTokenBucketLazy(1_000_000_000, 1_000_000, time.Millisecond)
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			l.mu.Lock()
			if l.bucket > 0 {
				l.bucket--
			}
			l.mu.Unlock()
		}
	})
}
