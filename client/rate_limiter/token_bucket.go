package rate_limiter_client

import (
	"fmt"
	"system_design/lib/rate_limiter/token_bucket"
	"time"
)

/*
Expected Output:

	Accept: 1, 2, 3, 4, 5
	Deny: 6, 7

	Sleep 1s

	Accept: 1, 2
	Deny: 3

	Sleep 2s

	Accept: 1, 2, 3, 4
	Deny: 5, 6, 7
*/
func TokenBucketBurst(tb token_bucket.TokenBucket) {
	for i := 1; i <= 7; i++ {
		fmt.Printf("Requesting: %d, %s\n", i, time.Now())
		if tb.Allow() {
			task(i)
		} else {
			deny(i)
		}
	}

	fmt.Println("Sleeping for 1 second...")
	fmt.Println("")
	time.Sleep(time.Second)
	time.Sleep(5 * time.Millisecond)

	for i := 1; i <= 3; i++ {
		fmt.Printf("Requesting: %d, %s\n", i, time.Now())
		if tb.Allow() {
			task(i)
		} else {
			deny(i)
		}
	}

	fmt.Println("Sleeping for 2 second...")
	fmt.Println("")
	time.Sleep(2 * time.Second)
	time.Sleep(5 * time.Millisecond)

	for i := 1; i <= 7; i++ {
		fmt.Printf("Requesting: %d, %s\n", i, time.Now())
		if tb.Allow() {
			task(i)
		} else {
			deny(i)
		}
	}
}
