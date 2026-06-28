package main

import (
	"fmt"
	"time"
)

// Ratelimiter limits processing to N items per second using channels only.
//
// Conceptually:
// - `bucket` is a token bucket (each token allows processing one item)
// - A ticker refills up to N tokens every second
// - Processing blocks until a token is available
// - Clean shutdown is handled explicitly to avoid leaks and panics
func rateLimiter(stream chan int, size int) {
	bucket := make(chan struct{}, size)

	// ✅ Pre-fill bucket
	for i := 0; i < size; i++ {
		bucket <- struct{}{}
	}

	// Refill tokens
	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for range ticker.C {
			for i := 0; i < size; i++ {
				select {
				case bucket <- struct{}{}:
				default:
				}
			}
		}
	}()

	for val := range stream {
		<-bucket // wait for token
		fmt.Println(val)
	}
}

func main() {
	input := make(chan int)

	// Start the rate limiter (5 items per second).
	go ratelimiter(input, 5)

	// Send some data to be rate-limited.
	for i := 0; i < 18; i++ {
		input <- i
	}

	// Close input to signal completion.
	close(input)
}
