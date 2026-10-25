// Online Go compiler to run Golang program online
// Print "Start small. Ship something." message

package main

import (
	"fmt"
	// "sort"
)

func searchInRotatedSortedArray(input []int, target int) int {

	left := 0
	right := len(input) - 1

	for left <= right {

		mid := left + (right-left)/2

		if input[mid] == target {
			return mid
		}

		// Left half is sorted
		if input[left] <= input[mid] {

			if input[left] <= target && target < input[mid] {
				right = mid - 1
			} else {
				left = mid + 1
			}

		} else {

			// Right half is sorted
			if input[mid] < target && target <= input[right] {
				left = mid + 1
			} else {
				right = mid - 1
			}
		}
	}

	return -1
}

func main() {

	input := []int{3, 4, 5, 7, 0, 1, 2}
	target := 2

	fmt.Println(searchInRotatedSortedArray(input, target))
}
