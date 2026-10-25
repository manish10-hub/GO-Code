//Given an unsorted array, find the length of the longest sequence of consecutive numbers.

/* Example
nums = [100, 4, 200, 1, 3, 2]
Answer = 4
The longest consecutive elements sequence is [1, 2, 3, 4]. Therefore its length is 4.
*/

// Pattern: Hash Map

package main

import "fmt"

func longestConsecutive(nums []int) int {
	numSet := make(map[int]bool)

	for _, num := range nums {
		numSet[num] = true
	}

	longest := 0

	for _, num := range nums {
		// Start only if num is the beginning
		// of a sequence.
		if !numSet[num-1] {
			current := num
			length := 1

			for numSet[current+1] {
				current++
				length++
			}

			if length > longest {
				longest = length
			}
		}
	}

	return longest
}

func main() {
	nums := []int{100, 4, 200, 1, 3, 2}

	fmt.Println(longestConsecutive(nums))
}
