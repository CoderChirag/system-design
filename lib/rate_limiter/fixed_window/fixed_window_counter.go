package fixed_window

import (
	"sync"
	"time"
)

type FixedWindowCounter struct {
	windowInterval      time.Duration
	counter, windowSize int
	lastWindow          time.Time
	mu                  sync.Mutex
}

func NewFixedWindowCounter(windowSize int, windowInterval time.Duration) *FixedWindowCounter {
	fwc := &FixedWindowCounter{
		windowSize:     windowSize,
		windowInterval: windowInterval,
		counter:        0,
		lastWindow:     time.Now(),
		mu:             sync.Mutex{},
	}
	return fwc
}

func (fwc *FixedWindowCounter) Allow() bool {
	fwc.mu.Lock()
	defer fwc.mu.Unlock()
	if n := int(time.Now().Sub(fwc.lastWindow) / fwc.windowInterval); n > 0 {
		fwc.counter = 1
		fwc.lastWindow = fwc.lastWindow.Add(time.Duration(n) * fwc.windowInterval)
		return true
	} else if fwc.counter < fwc.windowSize {
		fwc.counter++
		return true
	}
	return false
}
