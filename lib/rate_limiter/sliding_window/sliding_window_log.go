package sliding_window

import (
	"sync"
	"time"
)

type SlidingWindowLog struct {
	windowSize     int
	windowInterval time.Duration
	log            []time.Time
	mu             sync.Mutex
}

func NewSlidingWindow(windowSize int, windowInterval time.Duration) *SlidingWindowLog {
	sw := &SlidingWindowLog{
		windowSize:     windowSize,
		windowInterval: windowInterval,
		log:            []time.Time{},
		mu:             sync.Mutex{},
	}
	return sw
}

func (sw *SlidingWindowLog) Allow() bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	now := time.Now()
	i := 0
	for i < len(sw.log) && now.Sub(sw.log[i]) > sw.windowInterval {
		i++
	}
	sw.log = sw.log[i:]
	if len(sw.log) < sw.windowSize {
		sw.log = append(sw.log, now)
		return true
	}
	return false
}
