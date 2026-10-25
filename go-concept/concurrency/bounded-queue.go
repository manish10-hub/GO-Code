package main

import (
	"fmt"
)

type BoundedQueue struct {
	ch chan int
}

func NewBoundedQueue(capacity int) *BoundedQueue {
	return &BoundedQueue{
		ch: make(chan int, capacity),
	}
}

func (q *BoundedQueue) Enqueue(value int) {
	q.ch <- value
}

func (q *BoundedQueue) Dequeue() int {
	return <-q.ch
}

func main() {
	queue := NewBoundedQueue(3)

	queue.Enqueue(10)
	queue.Enqueue(20)
	queue.Enqueue(30)

	fmt.Println(queue.Dequeue())
	fmt.Println(queue.Dequeue())
	fmt.Println(queue.Dequeue())
}
