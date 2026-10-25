package main

import "fmt"

func findFirst(nums []int, target int) int {
	left := 0
	right := len(nums) - 1
	result := -1

	for left <= right {
		mid := left + (right-left)/2

		if nums[mid] == target {
			result = mid

			// We found target, but there may be
			// another occurrence on the left.
			right = mid - 1

		} else if nums[mid] < target {
			left = mid + 1

		} else {
			right = mid - 1
		}
	}

	return result
}

func findLast(nums []int, target int) int {
	left := 0
	right := len(nums) - 1
	result := -1

	for left <= right {
		mid := left + (right-left)/2

		if nums[mid] == target {
			result = mid

			// We found target, but there may be
			// another occurrence on the right.
			left = mid + 1

		} else if nums[mid] < target {
			left = mid + 1

		} else {
			right = mid - 1
		}
	}

	return result
}

func firstAndLastPosition(nums []int, target int) (int, int) {
	first := findFirst(nums, target)

	// Target doesn't exist.
	if first == -1 {
		return -1, -1
	}

	last := findLast(nums, target)

	return first, last
}

func main() {
	input := []int{5, 7, 7, 8, 8, 10}
	target := 8

	first, last := firstAndLastPosition(input, target)

	fmt.Println(first, last)
}
