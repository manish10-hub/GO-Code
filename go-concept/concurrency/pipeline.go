package main

import (
	"fmt"
	"sync"
)

func generate(nums []int) <-chan int {
	out := make(chan int, len(nums))
	go func() {
		for _, num := range nums {
			out <- num
		}
		close(out)
	}()
	return out
}

func square(nums <-chan int) <-chan int {
	out := make(chan int, 10)
	go func() {
		for num := range nums {
			out <- num * num
		}
		close(out)
	}()
	return out
}

func multiplyByTen(nums <-chan int) <-chan int {
	out := make(chan int, 10)
	go func() {
		for num := range nums {
			out <- num * 10
		}
		close(out)
	}()
	return out
}

func merge(channels ...<-chan int) <-chan int {
	out := make(chan int, 10)
	var wg sync.WaitGroup
	wg.Add(len(channels))

	for _, ch := range channels {
		go func(c <-chan int) {
			defer wg.Done()
			for num := range c {
				out <- num
			}
		}(ch)
	}

	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func main() {
	task := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}

	// Stage 1: Generate numbers
	firstCh := generate(task)

	// Stage 2: Square concurrently (fan-out)
	secondCh1 := square(firstCh)
	secondCh2 := square(firstCh)

	// Stage 3: Merge results (fan-in)
	mergedCh := merge(secondCh1, secondCh2)

	// Stage 4: Multiply concurrently
	forthCh1 := multiplyByTen(mergedCh)
	forthCh2 := multiplyByTen(mergedCh)
	forthCh3 := multiplyByTen(mergedCh)

	// Stage 5: Merge final output
	finalCh := merge(forthCh1, forthCh2, forthCh3)

	for num := range finalCh {
		fmt.Println("Final value:", num)
	}
}
