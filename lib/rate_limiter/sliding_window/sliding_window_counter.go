package sliding_window

import (
	"math"
	"sync"
	"time"
)

type SlidingWindowCounter struct {
	windowSize           int
	windowInterval       time.Duration
	counter, prevCounter int
	lastWindow           time.Time
	mu                   sync.Mutex
}

func NewSlidingWindowCounter(windowSize int, windowInterval time.Duration) *SlidingWindowCounter {
	sw := &SlidingWindowCounter{
		windowSize:     windowSize,
		windowInterval: windowInterval,
		counter:        0,
		prevCounter:    0,
		lastWindow:     time.Now(),
	}
	return sw
}

func (sw *SlidingWindowCounter) Allow() bool {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	windowDiff := time.Now().Sub(sw.lastWindow)
	if n := int(windowDiff / sw.windowInterval); n > 0 {
		sw.lastWindow = sw.lastWindow.Add(time.Duration(n) * sw.windowInterval)
		if n == 1 {
			sw.prevCounter = sw.counter
		} else {
			sw.prevCounter = 0
		}
		sw.counter = 0
	}
	prevOverlap := 1 - (float64(windowDiff%sw.windowInterval) / float64(sw.windowInterval))
	currReqs := int(math.Floor(float64(sw.counter) + float64(sw.prevCounter)*prevOverlap))
	if currReqs < sw.windowSize {
		sw.counter++
		return true
	}
	return false
}
