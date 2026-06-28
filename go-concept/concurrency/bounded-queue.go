package concurrency

import (
	"context"
	"sync"
)

type BoundedQueue struct {
	capacity int
	items    []interface{}
	head     int
	tail     int
	count    int
	lock     sync.Mutex
	notFull  *sync.Cond
	notEmpty *sync.Cond
}

func NewBoundedQueue(capacity int) *BoundedQueue {
	bq := &BoundedQueue{
		capacity: capacity,
		items:    make([]interface{}, capacity),
	}
	bq.notFull = sync.NewCond(&bq.lock)
	bq.notEmpty = sync.NewCond(&bq.lock)
	return bq
}

func (bq *BoundedQueue) Enqueue(ctx context.Context, item interface{}) bool {
	bq.lock.Lock()
	defer bq.lock.Unlock()

	for bq.count == bq.capacity {
		// Check context status inside loop to prevent lock sleep leaks
		if ctx.Err() != nil {
			return false
		}
		bq.notFull.Wait()
	}

	bq.items[bq.tail] = item
	bq.tail = (bq.tail + 1) % bq.capacity
	bq.count++

	bq.notEmpty.Signal()
	return true
}

func (bq *BoundedQueue) Dequeue(ctx context.Context) (interface{}, bool) {
	bq.lock.Lock()
	defer bq.lock.Unlock()

	for bq.count == 0 {
		if ctx.Err() != nil {
			return nil, false
		}
		bq.notEmpty.Wait()
	}

	item := bq.items[bq.head]
	bq.items[bq.head] = nil // Avoid memory leaks
	bq.head = (bq.head + 1) % bq.capacity
	bq.count--

	bq.notFull.Signal()
	return item, true
}
