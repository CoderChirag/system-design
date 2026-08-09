package rate_limiter_client

import (
	"fmt"
	"system_design/lib/rate_limiter/sliding_window"
	"time"
)

/*
Limiter under test: windowSize = 5, windowInterval = 1s.

Unlike the fixed window counter, the sliding window counter carries a *weighted*
share of the previous window's count into the current one:

	currReqs = floor(counter + prevCounter*(1 - elapsedFractionOfCurrentWindow))

So right after a saturated window the limiter is still (almost) full, and capacity
is released gradually as the previous window's weight decays. This burst walks
through window 2 at 10%, 50% and 90% elapsed to show that trickle, and then idles
for more than one full window to show prevCounter being dropped entirely.

Expected Output:

	t ~ 0.00s (window 1, prev = 0)
	  Accept: 1, 2, 3, 4, 5      counter = 5
	  Deny:   6, 7

	t ~ 1.10s (window 2, 10% elapsed, prev = 5 -> weight 5*0.90 = 4.5)
	  Accept: 8                  counter = 1  (4.5 + 0 = 4 < 5)
	  Deny:   9, 10              (4.5 + 1 = 5)

	t ~ 1.50s (window 2, 50% elapsed, prev weight = 2.5)
	  Accept: 11, 12             counter = 3  (2.5 + 1 = 3, 2.5 + 2 = 4)
	  Deny:   13                 (2.5 + 3 = 5)

	t ~ 1.90s (window 2, 90% elapsed, prev weight = 0.5)
	  Accept: 14, 15             counter = 5  (0.5 + 3 = 3, 0.5 + 4 = 4)
	  Deny:   16                 (0.5 + 5 = 5)

	=> only 5 requests served across the whole of window 2, but spread out
	   instead of being granted all at once at the window boundary.

	t ~ 4.10s (idle for more than one full window -> prev dropped to 0)
	  Accept: 17, 18, 19, 20, 21
	  Deny:   22, 23
*/
func SlidingWindowCounterBurst(sw *sliding_window.SlidingWindowCounter) {
	start := time.Now()

	fmt.Println("Window 1: bursting 7 requests...")
	fmt.Println("")
	sendSlidingWindowCounterBurst(sw, 1, 7)

	fmt.Println("Sleeping till 10% into window 2 (previous window still weighs 4.5)...")
	fmt.Println("")
	sleepUntil(start, 1100*time.Millisecond)
	sendSlidingWindowCounterBurst(sw, 8, 10)

	fmt.Println("Sleeping till 50% into window 2 (previous window now weighs 2.5)...")
	fmt.Println("")
	sleepUntil(start, 1500*time.Millisecond)
	sendSlidingWindowCounterBurst(sw, 11, 13)

	fmt.Println("Sleeping till 90% into window 2 (previous window now weighs 0.5)...")
	fmt.Println("")
	sleepUntil(start, 1900*time.Millisecond)
	sendSlidingWindowCounterBurst(sw, 14, 16)

	fmt.Println("Idling for more than a full window (previous count is dropped)...")
	fmt.Println("")
	sleepUntil(start, 4100*time.Millisecond)
	sendSlidingWindowCounterBurst(sw, 17, 23)
}

func sendSlidingWindowCounterBurst(sw *sliding_window.SlidingWindowCounter, from, to int) {
	for i := from; i <= to; i++ {
		fmt.Printf("Requesting: %d, %s\n", i, time.Now())
		if sw.Allow() {
			task(i)
		} else {
			deny(i)
		}
	}
}

// sleepUntil sleeps till start+offset instead of sleeping for a fixed duration, so
// that the time spent printing earlier bursts does not drift the window boundaries.
func sleepUntil(start time.Time, offset time.Duration) {
	if d := time.Until(start.Add(offset)); d > 0 {
		time.Sleep(d)
	}
}
