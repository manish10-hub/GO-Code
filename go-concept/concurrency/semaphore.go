package main

import (
	"fmt"
	"sync"
	"time"
)

func worker(id, val int, wg *sync.WaitGroup, sem chan struct{}) {
	defer wg.Done()
	sem <- struct{}{}        // acquire
	defer func() { <-sem }() // release

	fmt.Printf("Worker %d processing value %d\n", id, val)
	time.Sleep(1 * time.Second)
}

func main() {
	sem := make(chan struct{}, 3) // allow max 3 concurrent workers
	var wg sync.WaitGroup

	for i := 1; i <= 10; i++ {
		wg.Add(1)
		go worker(i, i*10, &wg, sem)
	}

	wg.Wait()
	fmt.Println("All workers done")
}
