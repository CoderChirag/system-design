package leaking_bucket

import (
	"sync"
	"time"
)

type task struct {
	fn func(id int)
	id int
}

type queue struct {
	q                                      []task
	length, startIdx, endIdx, outflowCount int
	lastOutflowWindow                      time.Time
}

type LeakingBucket struct {
	queueSize, outflowCount int
	outflowInterval         time.Duration
	q                       queue
	stop                    chan bool
	wg                      sync.WaitGroup
	mu                      sync.Mutex
}

func NewLeakingBucket(queueSize, outflowCount int, outflowInterval time.Duration) *LeakingBucket {
	lb := &LeakingBucket{
		queueSize:       queueSize,
		outflowCount:    outflowCount,
		outflowInterval: outflowInterval,
		q: queue{
			q:                 make([]task, queueSize),
			length:            0,
			startIdx:          0,
			endIdx:            0,
			outflowCount:      0,
			lastOutflowWindow: time.Now(),
		},
		stop: make(chan bool),
		wg:   sync.WaitGroup{},
		mu:   sync.Mutex{},
	}
	lb.startQueue()
	return lb
}

func (lb *LeakingBucket) startQueue() {
	go func() {
		for {
			select {
			case <-lb.stop:
				return
			default:
				lb.mu.Lock()
				if lb.q.length > 0 {
					if n := int(time.Now().Sub(lb.q.lastOutflowWindow) / lb.outflowInterval); n > 0 {
						task := lb.q.q[lb.q.startIdx] // intentionally copied inside critical section
						lb.q.length--
						lb.mu.Unlock()
						lb.q.outflowCount = 1
						lb.q.startIdx = (lb.q.startIdx + 1) % lb.queueSize
						lb.q.lastOutflowWindow = lb.q.lastOutflowWindow.Add(time.Duration(n) * lb.outflowInterval)
						go func() {
							defer lb.wg.Done()
							task.fn(task.id)
						}()
					} else if lb.q.outflowCount < lb.outflowCount {
						task := lb.q.q[lb.q.startIdx] // intentionally copied inside critical section
						lb.q.length--
						lb.mu.Unlock()
						lb.q.outflowCount++
						lb.q.startIdx = (lb.q.startIdx + 1) % lb.queueSize
						go func() {
							defer lb.wg.Done()
							task.fn(task.id)
						}()
					} else {
						lb.mu.Unlock()
					}
				} else {
					lb.mu.Unlock()
				}
			}
		}
	}()
}

func (lb *LeakingBucket) StopAndWait() {
	lb.wg.Wait()
	close(lb.stop)
}

func (lb *LeakingBucket) Request(fn func(id int), id int) bool {
	lb.mu.Lock()
	defer lb.mu.Unlock()
	if lb.q.length < lb.queueSize {
		lb.q.q[lb.q.endIdx] = task{
			fn: fn,
			id: id,
		}
		lb.wg.Add(1)
		lb.q.endIdx = (lb.q.endIdx + 1) % lb.queueSize
		lb.q.length++
		return true
	}
	return false
}
