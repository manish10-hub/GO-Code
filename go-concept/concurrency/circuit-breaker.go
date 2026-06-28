package main

import (
	"errors"
	"fmt"
	"sync"
	"time"
)

// State and CircuitBreaker definitions as provided
type State int

const (
	Closed State = iota
	Open
	HalfOpen
)

var ErrCircuitOpen = errors.New("circuit breaker is open")

type CircuitBreaker struct {
	mu              sync.RWMutex
	state           State
	failureCount    int
	threshold       int
	timeout         time.Duration
	lastStateChange time.Time
}

func NewCircuitBreaker(threshold int, timeout time.Duration) *CircuitBreaker {
	return &CircuitBreaker{
		state:     Closed,
		threshold: threshold,
		timeout:   timeout,
	}
}

func (cb *CircuitBreaker) Execute(req func() error) error {
	cb.mu.Lock()
	if cb.state == Open {
		if time.Since(cb.lastStateChange) > cb.timeout {
			cb.state = HalfOpen
			cb.lastStateChange = time.Now()
		} else {
			cb.mu.Unlock()
			return ErrCircuitOpen
		}
	}
	cb.mu.Unlock()

	err := req()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.failureCount++
		if cb.failureCount >= cb.threshold || cb.state == HalfOpen {
			cb.state = Open
			cb.lastStateChange = time.Now()
		}
		return err
	}

	if cb.state == HalfOpen {
		cb.state = Closed
		cb.failureCount = 0
	}
	return nil
}

// Helper function to print execution statuses
func runRequest(cb *CircuitBreaker, id int, simulateErr bool) {
	mockRequest := func() error {
		if simulateErr {
			return errors.New("downstream service timeout error")
		}
		return nil
	}

	err := cb.Execute(mockRequest)
	if err != nil {
		if errors.Is(err, ErrCircuitOpen) {
			fmt.Printf("Request #%d: 🔴 Blocked by Circuit Breaker (%v)\n", id, err)
		} else {
			fmt.Printf("Request #%d: ⚠️ Request Failed (%v), failure count updating\n", id, err)
		}
	} else {
		fmt.Printf("Request #%d: 🟢 Request Succeeded!\n", id)
	}
}

func main() {
	// 1. Initialize Breaker: Trip after 2 errors, cool down for 2 seconds
	threshold := 2
	cooldown := 2 * time.Second
	cb := NewCircuitBreaker(threshold, cooldown)

	fmt.Println("--- STAGE 1: Breaker is CLOSED (Normal Behavior) ---")
	runRequest(cb, 1, false) // Success
	runRequest(cb, 2, false) // Success

	fmt.Println("\n--- STAGE 2: Tripping the Breaker ---")
	runRequest(cb, 3, true) // Fail 1
	runRequest(cb, 4, true) // Fail 2 -> Exceeds threshold, trips to OPEN state

	fmt.Println("\n--- STAGE 3: Breaker is OPEN (Fails Fast) ---")
	// Subsequent requests should be intercepted immediately without invoking the lambda
	runRequest(cb, 5, false)
	runRequest(cb, 6, false)

	fmt.Println("\n--- STAGE 4: Waiting for cooldown timeout to expire... ---")
	time.Sleep(2500 * time.Millisecond)

	fmt.Println("\n--- STAGE 5: Breaker enters HALF-OPEN (Testing Canary Request) ---")
	// The first request after the cooldown window moves the internal state machine to Half-Open.
	// If this request fails, the breaker flips right back to Open. Let's send a successful request:
	runRequest(cb, 7, false)

	fmt.Println("\n--- STAGE 6: Breaker resets back to CLOSED ---")
	// Because request #7 succeeded in Half-Open, the breaker closed down again to accept full traffic.
	runRequest(cb, 8, false)
	runRequest(cb, 9, false)
}
