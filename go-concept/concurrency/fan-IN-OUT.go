package main

import (
	"fmt"
	"sync"
)

// producer generates data and sends it into its output channel.
// Each producer sends 3 values: (id*10 + 0), (id*10 + 1), (id*10 + 2).
// Once done, it closes its channel.
func producer(id int, out chan int) {
	for i := 0; i < 3; i++ {
		out <- (id*10 + i)
	}
	close(out)
}

// fanIn merges multiple input channels (chanArr) into a single output channel (result).
// It waits for all input channels to close before closing the result channel.
// This is the FAN-IN pattern.
func fanIn(chanArr []chan int) chan int {
	result := make(chan int)
	var wg sync.WaitGroup
	wg.Add(len(chanArr))

	for _, ch := range chanArr {
		go func(c chan int) {
			defer wg.Done()
			for currCh := range c {
				// Forward each value from producer channels into the merged result channel
				result <- currCh
			}
		}(ch)
	}

	// Once all producer channels are done, close the merged result channel
	go func() {
		wg.Wait()
		close(result)
	}()

	return result
}

// fanOut consumes data from a shared input channel (dataStream) concurrently.
// Multiple fanOut workers (goroutines) read from the same dataStream channel.
// This is the FAN-OUT pattern.
func fanOut(id int, dataStream chan int, wg *sync.WaitGroup) {
	defer wg.Done()
	for data := range dataStream {
		fmt.Println("consuming from worker", id, "value:", data)
	}
}

func main() {
	totalProducer := 5
	chanArr := make([]chan int, totalProducer)
	dataStream := make(chan int)

	// Start multiple producers — each writes its own channel
	for i := 0; i < totalProducer; i++ {
		chanArr[i] = make(chan int)
		go producer(i, chanArr[i])
	}

	// Merge all producer channels into one
	merged := fanIn(chanArr)

	// Start multiple consumer workers — each reads from shared dataStream
	workers := 4
	var wg sync.WaitGroup
	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go fanOut(i, dataStream, &wg)
	}

	// Send merged data to shared dataStream channel
	for data := range merged {
		dataStream <- data
	}

	// Close the dataStream channel once all data is sent
	close(dataStream)

	// Wait until all consumers finish
	wg.Wait()
}

/*
Visualization of the Flow

         +-----------+
         | Producer 0 | → ch0
         +-----------+
               |
         +-----------+
         | Producer 1 | → ch1
         +-----------+
               |
         +-----------+
         | Producer 2 | → ch2
         +-----------+
               |
         +-----------+
         | Producer 3 | → ch3
         +-----------+
               |
         +-----------+
         | Producer 4 | → ch4
         +-----------+
               |
               ↓
   =========================
   ||     fanIn (merge)   ||
   =========================
               ↓
         merged channel
               ↓
   =========================
   ||   dataStream channel ||
   =========================
               ↓
      +------------------+
      |  fanOut worker 0 |
      +------------------+
      |  fanOut worker 1 |
      +------------------+
      |  fanOut worker 2 |
      +------------------+
      |  fanOut worker 3 |
      +------------------+

Each worker reads from the same dataStream channel concurrently
until it’s closed.

*/
