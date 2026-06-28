package main

import (
	"fmt"
	"sync"
)

type Job struct {
	ID int
}

func worker(id int, jobs <-chan Job, wg *sync.WaitGroup) {
	defer wg.Done()

	for job := range jobs {
		fmt.Printf("worker=%d processing job=%d\n", id, job.ID)
	}
}

func main() {
	const numWorkers = 3

	jobs := make(chan Job, 10)

	var wg sync.WaitGroup

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go worker(i, jobs, &wg)
	}

	for i := 0; i < 20; i++ {
		jobs <- Job{ID: i}
	}

	close(jobs)

	wg.Wait()
}
