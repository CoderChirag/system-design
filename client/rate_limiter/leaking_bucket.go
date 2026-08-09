package rate_limiter_client

import (
	"fmt"
	"system_design/lib/rate_limiter/leaking_bucket"
	"time"
)

/*
Expected Output:

	T0
	Accept: 1, 2, 3, 4, 5, 6, 7
	Respond: 1, 2

	T1
	Accept: 8, 9
	Deny: 10, 11, 12
	Respond: 3, 4

	T2
	Respond: 5, 6

	T3
	Accept: 13, 14, 15, 16
	Deny: 17, 18, 19, 20
	Respond: 7, 8

	T4
	Respond: 9, 13

	T5
	Respond: 14, 15

	T6
	Respond: 16
*/
func LeakingBucketburst(lb *leaking_bucket.LeakingBucket) {
	for i := 1; i <= 7; i++ {
		fmt.Printf("Requesting: %d, %s\n", i, time.Now())
		if !lb.Request(task, i) {
			deny(i)
		}
	}

	fmt.Println("Sleeping for 1 second...")
	fmt.Println("")
	time.Sleep(time.Second)
	time.Sleep(5 * time.Millisecond)

	for i := 8; i <= 12; i++ {
		fmt.Printf("Requesting: %d, %s\n", i, time.Now())
		if !lb.Request(task, i) {
			deny(i)
		}
	}

	fmt.Println("Sleeping for 2 second...")
	fmt.Println("")
	time.Sleep(2 * time.Second)
	time.Sleep(5 * time.Millisecond)

	for i := 13; i <= 20; i++ {
		fmt.Printf("Requesting: %d, %s\n", i, time.Now())
		if !lb.Request(task, i) {
			deny(i)
		}
	}
}
