package main

// IntQueue represents a queue of integers.
type IntQueue struct {
	data []int
}

// Enqueue adds an element to the end of the queue.
func (q *IntQueue) Enqueue(value int) {
	q.data = append(q.data, value)
}

// Dequeue removes and returns the element at the front of the queue.
func (q *IntQueue) Dequeue() (int, bool) {
	if len(q.data) == 0 {
		return 0, false
	}
	val := q.data[0]
	q.data = q.data[1:] // remove the front element
	return val, true
}

// Peek returns the front element without removing it.
func (q *IntQueue) Peek() (int, bool) {
	if len(q.data) == 0 {
		return 0, false
	}
	return q.data[0], true
}

// IsEmpty checks if the queue is empty.
func (q *IntQueue) IsEmpty() bool {
	return len(q.data) == 0
}
