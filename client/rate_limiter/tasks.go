package rate_limiter_client

import (
	"fmt"
	"time"
)

func task(id int) {
	fmt.Printf("[200] Request Served: %d, %s\n", id, time.Now())
	fmt.Println("")
}

func deny(id int) {
	fmt.Printf("[429] Too Many Requests: %d, %s\n\n", id, time.Now())
}
