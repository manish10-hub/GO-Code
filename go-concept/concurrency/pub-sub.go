package main

import (
	"fmt"
	"sync"
	"time"
)

// Message and PubSub definitions as provided
type Message struct {
	Topic string
	Data  interface{}
}

type PubSub struct {
	mu          sync.RWMutex
	subscribers map[string][]chan Message
	closed      bool
}

func NewPubSub() *PubSub {
	return &PubSub{
		subscribers: make(map[string][]chan Message),
	}
}

func (ps *PubSub) Subscribe(topic string, bufferSize int) chan Message {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	ch := make(chan Message, bufferSize)
	ps.subscribers[topic] = append(ps.subscribers[topic], ch)
	return ch
}

func (ps *PubSub) Publish(topic string, data interface{}) {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	if ps.closed {
		return
	}

	msg := Message{Topic: topic, Data: data}
	for _, ch := range ps.subscribers[topic] {
		select {
		case ch <- msg:
		case <-time.After(time.Millisecond * 10): // Prevent slower receivers from blocking the broker
		}
	}
}

func (ps *PubSub) Close() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	if ps.closed {
		return
	}
	ps.closed = true

	for _, topicSubs := range ps.subscribers {
		for _, ch := range topicSubs {
			close(ch)
		}
	}
}

func main() {
	// 1. Initialize the PubSub broker
	broker := NewPubSub()
	var wg sync.WaitGroup

	// 2. Create Subscribers for the "news" topic
	newsSub1 := broker.Subscribe("news", 5)
	newsSub2 := broker.Subscribe("news", 5)

	// 3. Create a Subscriber for the "sports" topic
	sportsSub := broker.Subscribe("sports", 5)

	// 4. Create an intentionally SLOW Subscriber (buffer size 0) to test non-blocking protection
	slowSub := broker.Subscribe("news", 0)

	// Launch Goroutine for News Subscriber 1
	wg.Add(1)
	go func() {
		defer wg.Done()
		for msg := range newsSub1 {
			fmt.Printf("[News Sub 1] Received: %v\n", msg.Data)
		}
		fmt.Println("[News Sub 1] Channel closed, exiting.")
	}()

	// Launch Goroutine for News Subscriber 2
	wg.Add(1)
	go func() {
		defer wg.Done()
		for msg := range newsSub2 {
			fmt.Printf("[News Sub 2] Received: %v\n", msg.Data)
		}
		fmt.Println("[News Sub 2] Channel closed, exiting.")
	}()

	// Launch Goroutine for Sports Subscriber
	wg.Add(1)
	go func() {
		defer wg.Done()
		for msg := range sportsSub {
			fmt.Printf("[Sports Sub] Received: %v\n", msg.Data)
		}
		fmt.Println("[Sports Sub] Channel closed, exiting.")
	}()

	// Launch Goroutine for Slow Subscriber
	wg.Add(1)
	go func() {
		defer wg.Done()
		// Sleep initially so it misses the first broadcast completely
		time.Sleep(50 * time.Millisecond)
		for msg := range slowSub {
			fmt.Printf("[Slow Sub] Received: %v\n", msg.Data)
		}
		fmt.Println("[Slow Sub] Channel closed, exiting.")
	}()

	// 5. Publish events to topics
	fmt.Println("[Broker] Publishing updates...")
	broker.Publish("news", "Breaking: Go 1.27 released!")
	broker.Publish("sports", "Score: Team A wins 3-2!")
	broker.Publish("news", "Tech: Concurrency patterns are powerful.")

	// Give subscribers a brief moment to process incoming messages
	time.Sleep(100 * time.Millisecond)

	// 6. Tear down the broker cleanly
	fmt.Println("\n[Broker] Shutting down broker...")
	broker.Close()

	// Wait for all reader goroutines to finish after channels close
	wg.Wait()
	fmt.Println("[Main] All workers stopped cleanly.")
}
