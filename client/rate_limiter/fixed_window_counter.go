package rate_limiter_client

import (
	"fmt"
	"system_design/lib/rate_limiter/fixed_window"
	"time"
)

/*
Expected Output:

	Accept: 1, 2, 3, 4, 5
	Deny: 6, 7

	Sleep 1.995s

	Accept: 8, 9, 10, 11, 12, 13, 14, 15
*/
func FixedWindowCounterBurst(tb *fixed_window.FixedWindowCounter) {
	for i := 1; i <= 7; i++ {
		fmt.Printf("Requesting: %d, %s\n", i, time.Now())
		if tb.Allow() {
			task(i)
		} else {
			deny(i)
		}
	}

	fmt.Println("Sleeping for 1.995 second...")
	fmt.Println("")
	time.Sleep(1995 * time.Millisecond)

	for i := 8; i <= 12; i++ {
		fmt.Printf("Requesting: %d, %s\n", i, time.Now())
		if tb.Allow() {
			task(i)
		} else {
			deny(i)
		}
	}

	fmt.Println("Edge of windows, more than 5 requests getting served...")
	fmt.Println("")
	time.Sleep(5 * time.Millisecond) // Minimal Sleep just for simulating the edge of Window

	for i := 13; i <= 15; i++ {
		fmt.Printf("Requesting: %d, %s\n", i, time.Now())
		if tb.Allow() {
			task(i)
		} else {
			deny(i)
		}
	}
}
