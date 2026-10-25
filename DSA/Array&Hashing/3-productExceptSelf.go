// Given an array, return an array where: answer[i] = product of every element except nums[i]

/*Example
Input: nums = [1,2,3,4]
Output: [24,12,8,6]
*/

// Pattern: Prefix and Suffix Product
package main

import "fmt"

func productExceptSelf(nums []int) []int {
	n := len(nums)

	answer := make([]int, n)

	// Prefix product
	prefix := 1

	for i := 0; i < n; i++ {
		answer[i] = prefix
		prefix *= nums[i]
	}

	// Suffix product
	suffix := 1

	for i := n - 1; i >= 0; i-- {
		answer[i] *= suffix
		suffix *= nums[i]
	}

	return answer
}

func main() {
	nums := []int{1, 2, 3, 4}

	fmt.Println(productExceptSelf(nums))
}
