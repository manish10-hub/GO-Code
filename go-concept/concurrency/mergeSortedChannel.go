package main

import "fmt"

func mergeSortedChannels(ch1, ch2 <-chan int) <-chan int {
	out := make(chan int)

	go func() {
		defer close(out)

		v1, ok1 := <-ch1
		v2, ok2 := <-ch2

		for ok1 && ok2 {
			if v1 <= v2 {
				out <- v1
				v1, ok1 = <-ch1
			} else {
				out <- v2
				v2, ok2 = <-ch2
			}
		}

		// Drain remaining values
		for ok1 {
			out <- v1
			v1, ok1 = <-ch1
		}
		for ok2 {
			out <- v2
			v2, ok2 = <-ch2
		}
	}()

	return out
}

func main() {

	ch1 := make(chan int)
	ch2 := make(chan int)

	go func() {
		ch1 <- 1
		ch1 <- 3
		ch1 <- 5
		close(ch1)
	}()

	go func() {
		ch2 <- 2
		ch2 <- 4
		ch2 <- 6
		close(ch2)
	}()

	merged := mergeSortedChannels(ch1, ch2)

	for v := range merged {
		fmt.Println(v)
	}
}
