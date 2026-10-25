//Given an array of integers and a target, find the two indices whose values add up to the target.
/*
nums = [2, 7, 11, 15]
target = 9

2 + 7 = 9

Answer = [0, 1]
*/

// Pattern: Hash Map

package main

import "fmt"

func twoSum(nums []int, target int) []int {
	seen := make(map[int]int)

	for i, num := range nums {
		required := target - num

		if index, found := seen[required]; found {
			return []int{index, i}
		}

		seen[num] = i
	}

	return []int{}
}

func main() {
	nums := []int{2, 7, 11, 15}

	fmt.Println(twoSum(nums, 9))
}
