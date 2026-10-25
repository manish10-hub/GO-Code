package main

import (
	"container/heap"
	"fmt"
)

type Item struct {
	value int
	count int
}

type MinHeap []Item

func (h MinHeap) Len() int {
	return len(h)
}

func (h MinHeap) Less(i, j int) bool {
	return h[i].count < h[j].count
}

func (h MinHeap) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
}

func (h *MinHeap) Push(x interface{}) {
	*h = append(*h, x.(Item))
}

func (h *MinHeap) Pop() interface{} {
	old := *h
	n := len(old)

	item := old[n-1]
	*h = old[:n-1]

	return item
}

func topKFrequent(nums []int, k int) []int {
	// Step 1: Count frequency
	freq := make(map[int]int)

	for _, num := range nums {
		freq[num]++
	}

	// Step 2: Keep only K most frequent elements
	h := &MinHeap{}
	heap.Init(h)

	for value, count := range freq {
		heap.Push(h, Item{
			value: value,
			count: count,
		})

		if h.Len() > k {
			heap.Pop(h)
		}
	}

	// Step 3: Extract result
	result := make([]int, h.Len())

	for i := len(result) - 1; i >= 0; i-- {
		result[i] = heap.Pop(h).(Item).value
	}

	return result
}

func main() {
	nums := []int{1, 1, 1, 2, 2, 3}
	k := 2

	fmt.Println(topKFrequent(nums, k))
}
